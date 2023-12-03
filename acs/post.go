package acs

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/netdoop/cwmp/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

func (s *AcsServer) HandlePost(c echo.Context) error {
	r := c.Request()
	contentType := r.Header.Get("Content-Type")
	ctx := r.Context()
	sess, _ := session.Get("session", c)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "read body").Error())
	}

	var msg *proto.SoapEnvelope = nil
	if len(body) > 0 {
		body = proto.CleanXMLData(body)
		msg = &proto.SoapEnvelope{}
		if err := xml.Unmarshal(body, msg); err != nil {
			s.logger.Error("handle post", zap.Error(err))
			s.logger.Debug("debug", zap.String("body", string(body)))
			return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unmarshal xml").Error())
		}
		msg.Decode()
	}
	//logger.Warnf("%v", msg.Attrs)
	//logger.Warnf("%v %v %v %v %v", msg.SoapEnv, msg.SoapEnc, msg.Xsd, msg.Xsi, msg.Cwmp)

	if msg != nil {
		if msg.Body.Inform != nil {
			// logger.Warn("debug", zap.String("inform", string(body)))
			if err := s.handleInform(ctx, msg.Body.Inform); err != nil {
				err = errors.Wrap(err, "handle Inform")
				msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
				s.logger.Error("handle post", zap.Error(err))
				c.Response().Header().Set("Content-Type", contentType)
				return c.XML(http.StatusOK, msg2)
			}
			sess.Options = &sessions.Options{
				Path:     "/acs",
				MaxAge:   1800,
				HttpOnly: true,
			}
			sess.Values["OUI"] = msg.Body.Inform.DeviceID.OUI
			sess.Values["ProductClass"] = msg.Body.Inform.DeviceID.ProductClass
			sess.Values["SerialNumber"] = msg.Body.Inform.DeviceID.SerialNumber
			sess.Values["SoapNamespace"] = msg.NS.ToString()
			sess.Values["ContentType"] = contentType
			sess.Save(c.Request(), c.Response())

			msg2 := proto.CreateEnvelope(msg.Header.ID.Text, msg.NS, &proto.InformResponse{MaxEnvelopes: 1})
			c.Response().Header().Set("Content-Type", contentType)
			return c.XML(http.StatusOK, msg2)
		}
	}
	var ns proto.SoapNamespace
	var hasNS bool = false
	if v, ok := sess.Values["SoapNamespace"]; ok && v != nil {
		if v2, ok := v.(string); ok {
			ns.FromString(v2)
			hasNS = true
		}
	}
	if !hasNS {
		return c.NoContent(http.StatusOK)
	}

	device := s.getDeviceBySession(sess)
	if device == nil {
		sess.Save(c.Request(), c.Response())
		err := errors.New("invalid session")
		s.logger.Error("handle post", zap.Error(err))
		msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
		return s.responseXML(c, sess, msg2)
	}
	product := device.Product()
	if product == nil {
		s.logger.Error("unknow product")
		msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
		return s.responseXML(c, sess, msg2)
	}

	if msg != nil {
		if msg.Body.TransferComplete != nil {
			if err := s.handleTransferComplete(ctx, device, msg.Body.TransferComplete); err != nil {
				err = errors.Wrap(err, "handle TransferComplete")
				s.logger.Error("handle post", zap.Error(err))
				msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
				return s.responseXML(c, sess, msg2)
			}
			msg2 := proto.CreateEnvelope(msg.Header.ID.Text, msg.NS, &proto.TransferCompleteResponse{})
			return s.responseXML(c, sess, msg2)
		}
		if msg.Body.AutonomousTransferComplete != nil {
			if err := s.handleAutonomousTransferComplete(ctx, device, msg.Body.AutonomousTransferComplete); err != nil {
				err = errors.Wrap(err, "handle AutonomousTransferComplete")
				s.logger.Error("handle post", zap.Error(err))
				msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
				return s.responseXML(c, sess, msg2)
			}
			msg2 := proto.CreateEnvelope(msg.Header.ID.Text, msg.NS, &proto.AutonomousTransferCompleteResponse{})
			return s.responseXML(c, sess, msg2)
		}
		if err2 := s.handleFault(device, msg.Header.ID.Text, msg.Body.Fault); err2 != nil {
			s.logger.Error("handle SoapFault", zap.Error(err2))
		}

		if err := s.handleGetRPCMethodsResponse(ctx, device, msg.Header.ID.Text, msg.Body.GetRPCMethodsResponse); err != nil {
			err = errors.Wrap(err, "handle GetRPCMethodsResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleGetParameterValuesResponse(ctx, device, msg.Header.ID.Text, msg.Body.GetParameterValuesResponse); err != nil {
			err = errors.Wrap(err, "handle GetParameterValuesResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleSetParameterValuesResponse(ctx, device, msg.Header.ID.Text, msg.Body.SetParameterValuesResponse); err != nil {
			err = errors.Wrap(err, "handle SetParameterValuesResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleGetParameterNamesResponse(ctx, device, msg.Header.ID.Text, msg.Body.GetParameterNamesResponse); err != nil {
			err = errors.Wrap(err, "handle GetParameterNamesResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleSetParameterAttributesResponse(ctx, device, msg.Header.ID.Text, msg.Body.SetParameterAttributesResponse); err != nil {
			err = errors.Wrap(err, "handle SetParameterAttributesResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleGetParameterAttributesResponse(ctx, device, msg.Header.ID.Text, msg.Body.GetParameterAttributesResponse); err != nil {
			err = errors.Wrap(err, "handle GetParameterAttributesResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleAddObjectResponse(ctx, device, msg.Header.ID.Text, msg.Body.AddObjectResponse); err != nil {
			err = errors.Wrap(err, "handle AddObjectResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleDeleteObjectResponse(ctx, device, msg.Header.ID.Text, msg.Body.DeleteObjectResponse); err != nil {
			err = errors.Wrap(err, "handle DeleteObjectResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleDownloadResponse(ctx, device, msg.Header.ID.Text, msg.Body.DownloadResponse); err != nil {
			err = errors.Wrap(err, "handle DownloadResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleUploadResponse(ctx, device, msg.Header.ID.Text, msg.Body.UploadResponse); err != nil {
			err = errors.Wrap(err, "handle UploadResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleRebootResponse(ctx, device, msg.Header.ID.Text, msg.Body.RebootResponse); err != nil {
			err = errors.Wrap(err, "handle RebootResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}
		if err := s.handleFactoryResetResponse(ctx, device, msg.Header.ID.Text, msg.Body.FactoryResetResponse); err != nil {
			err = errors.Wrap(err, "handle FactoryResetResponse")
			s.logger.Error("handle post", zap.Error(err))
			msg2 := proto.CreateEnvelopeFault(msg.Header.ID.Text, msg.NS, proto.ACSFaultCodeInternalError, err)
			return s.responseXML(c, sess, msg2)
		}

		if v := msg.Body.TransferComplete; v != nil {
			s.logger.Warn("TransferComplete")
		}
		if v := msg.Body.AutonomousTransferComplete; v != nil {
			s.logger.Warn("AutonomousTransferComplete")
		}
	}

	msg2 := s.getNextMessage(ns, device)
	if msg2 != nil {
		sess.Save(c.Request(), c.Response())
		return s.responseXML(c, sess, msg2)
	}
	return c.NoContent(http.StatusOK)
}

func (s *AcsServer) getNextMessage(ns proto.SoapNamespace, device Device) *proto.SoapEnvelope {
	product := device.Product()
	dm := product.DataModel()
	var body any = nil
	if call := device.GetNextMethodCall(); call != nil {
		commandKey := call.CommandKey()
		methodName := call.MethodName()
		values := call.RequestValues()
		if values != nil && len(values) > 0 {
			s.logger.Debug("device method",
				zap.String("method", methodName),
				zap.String("command_key", commandKey),
				zap.Int("requset_values", len(values)),
			)
		} else {
			s.logger.Debug("device method",
				zap.String("method", methodName),
				zap.String("command_key", commandKey),
			)
		}

		switch methodName {
		case "GetRPCMethods":
			body = &proto.GetRPCMethods{}
		case "GetParameterNames":
			path := call.GetRequestValue("ParameterPath")
			nextLevel := call.GetRequestValue("NextLevel")
			body = &proto.GetParameterNames{
				ParameterPath: path,
				NextLevel:     nextLevel,
			}
		case "GetParameterValues":
			params := proto.ParameterNames{}
			if values != nil && len(values) > 0 {
				for k := range values {
					params.Strings = append(params.Strings, k)
				}
			} else {
				params.Strings = device.GetParameterNames("", false)
			}
			body = &proto.GetParameterValues{
				ParameterNames: params,
			}
		case "SetParameterValues":
			params := []*proto.ParameterValueStruct{}
			if values != nil && len(values) > 0 {
				for k, v := range values {
					param := proto.ParameterValueStruct{
						Name:  proto.ParameterName{Text: k},
						Value: proto.ParameterValue{Text: cast.ToString(v), TypeName: "string"},
					}
					if typeName := dm.GetParameterType(k); typeName != "" {
						param.Value.TypeName = typeName
					}
					params = append(params, &param)
				}
			}
			tmp := &proto.SetParameterValues{}
			tmp.ParameterList.ParameterValueStructs = params
			body = tmp
		case "GetParameterAttributes":
			params := proto.ParameterNames{}
			if values != nil && len(values) > 0 {
				for k := range values {
					params.Strings = append(params.Strings, k)
				}
			} else {
				params.Strings = device.GetParameterNames("", false)
			}
			body = &proto.GetParameterAttributes{
				ParameterNames: params,
			}
		case "SetParameterAttributes":
			params := []*proto.SetParameterAttributesStruct{}
			if values != nil && len(values) > 0 {
				for k, v := range values {
					notification := cast.ToInt(v)
					if notification == 1 || notification == 2 {
						params = append(params, &proto.SetParameterAttributesStruct{
							Name:               k,
							NotificationChange: true,
							Notification:       notification,
						})
					}
				}
			}
			tmp := &proto.SetParameterAttributes{}
			tmp.ParameterList.SetParameterAttributesStructs = params
			body = tmp
		case "AddObject":
			objectName := call.GetRequestValue("ObjectName")
			parameterKey := call.GetRequestValue("ParameterKey")
			body = &proto.AddObject{
				ObjectName:   objectName,
				ParameterKey: parameterKey,
			}
		case "DeleteObject":
			objectName := call.GetRequestValue("ObjectName")
			parameterKey := call.GetRequestValue("ParameterKey")
			body = &proto.DeleteObject{
				ObjectName:   objectName,
				ParameterKey: parameterKey,
			}
		case "Download":
			url := call.GetRequestValue("Url")
			username := call.GetRequestValue("Username")
			password := call.GetRequestValue("Password")
			fileType := call.GetRequestValue("FileType")
			fileSize := cast.ToUint(call.GetRequestValue("FileSize"))
			targetFileName := call.GetRequestValue("TargetFileName")
			delaySeconds := cast.ToUint(call.GetRequestValue("DelaySeconds"))
			successURL := call.GetRequestValue("SuccessURL")
			failureURL := call.GetRequestValue("FailureURL")

			body = &proto.Download{
				CommandKey:     commandKey,
				URL:            url,
				Username:       username,
				Password:       password,
				FileType:       fileType,
				FileSize:       fileSize,
				TargetFileName: targetFileName,
				DelaySeconds:   delaySeconds,
				SuccessURL:     successURL,
				FailureURL:     failureURL,
			}
		case "Upload":
			url := call.GetRequestValue("Url")
			username := call.GetRequestValue("Username")
			password := call.GetRequestValue("Password")
			fileType := call.GetRequestValue("FileType")
			delaySeconds := cast.ToUint(call.GetRequestValue("DelaySeconds"))

			body = &proto.Upload{
				CommandKey:   commandKey,
				URL:          url,
				Username:     username,
				Password:     password,
				FileType:     fileType,
				DelaySeconds: delaySeconds,
			}
		case "Reboot":
			body = &proto.Reboot{
				CommandKey: commandKey,
			}
		case "FactoryReset":
			body = &proto.FactoryReset{}
		default:
			device.UpdateMethodCallUnknow(commandKey)
			s.logger.Warn("unsupported device method",
				zap.String("method", methodName),
				zap.String("command_key", commandKey),
			)
			return nil
		}
		if body != nil {
			device.UpdateMethodCallRequestSend(commandKey)
			return proto.CreateEnvelope(commandKey, ns, body)
		}
	}
	return nil
}

func (s *AcsServer) getDeviceBySession(sess *sessions.Session) Device {
	oui := cast.ToString(sess.Values["OUI"])
	productClass := cast.ToString(sess.Values["ProductClass"])
	serialNumber := cast.ToString(sess.Values["SerialNumber"])
	return s.deviceStore.GetDevice("", oui, productClass, serialNumber)
}

func (s *AcsServer) responseXML(c echo.Context, sess *sessions.Session, i any) error {
	if v, ok := sess.Values["ContentType"]; ok && v != nil {
		if v2, ok := v.(string); ok {
			c.Response().Header().Set("Content-Type", v2)
		}
	}
	return c.XML(http.StatusOK, i)
}
