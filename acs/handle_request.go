package acs

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/netdoop/cwmp/pm"
	"github.com/netdoop/cwmp/proto"

	"github.com/heypkg/store/jsontype"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

func (s *AcsServer) handleFault(device Device, id string, v *proto.SoapFault) error {
	if v == nil {
		return nil
	}
	s.logger.Debug("SoapFault", zap.Any("SoapFault", v))
	faultCode := 0
	faultString := ""
	values := jsontype.Tags{}
	if v.FaultCode == "Client" && v.FaultString == "CWMP Fault" {
		for _, v2 := range v.Detail.Fault.SetParameterValuesFaults {
			s.logger.Debug("SetParameterValuesFault",
				zap.Int("FaultCode", v2.FaultCode),
				zap.String("FaultString", v2.FaultString),
				zap.String("ParameterName", v2.ParameterName),
			)
			values[v2.ParameterName] = fmt.Sprintf("%d:%v", v2.FaultCode, v2.FaultString)
		}
		faultCode = cast.ToInt(v.Detail.Fault.FaultCode)
		faultString = v.Detail.Fault.FaultString
	} else {
		faultString = v.FaultString
	}
	device.UpdateMethodCallResponse(id, values, faultCode, faultString)
	return nil
}

func (s *AcsServer) handleInform(ctx context.Context, inform *proto.Inform) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	if inform == nil {
		return nil
	}
	s.logger.Debug("Inform", zap.Any("Inform", inform))
	product := s.GetProduct("", inform.DeviceID.OUI, inform.DeviceID.ProductClass)
	if product == nil {
		return errors.New("unknow product")
	}

	values := map[string]string{}
	for _, param := range inform.ParameterList.ParameterValueStructs {
		values[param.Name.Text] = param.Value.Text
	}

	lastOnlineStatus := false
	currentTime, _ := proto.ParseTime(inform.CurrentTime)
	periodicOnly := true
	needSync := false
	for _, event := range inform.Event.Events {
		if event.EventCode != "2 PERIODIC" && event.EventCode != "10 AUTONOMOUS TRANSFER COMPLETE" {
			periodicOnly = false
		}
		if event.EventCode == "1 BOOT" || event.EventCode == "0 BOOTSTRAP" {
			needSync = true
		}
	}

	schema := ""
	device := s.GetDevice(schema, inform.DeviceID.OUI, inform.DeviceID.ProductClass, inform.DeviceID.SerialNumber)
	if device == nil {
		device = s.deviceStore.CreateDeviceWithInform(schema, inform.DeviceID.OUI, inform.DeviceID.ProductClass, inform.DeviceID.SerialNumber, inform)
		device.PushMethodCall(time.Now(), "GetRPCMethods", nil)
	} else {
		device.UpdateParameterValues(values)
		lastOnlineStatus = device.OnlineStatus()
	}
	device.HandleAlive(currentTime, lastOnlineStatus)

	if needSync {
		device.GetDeviceParameterNames("Device.", false)
	}

	if !periodicOnly {
		metaData := map[string]any{}
		for _, v := range inform.Event.Events {
			metaData[v.EventCode] = v.CommandKey
		}
		if err := device.InsertEvent("Inform", proto.MustParseTime(inform.CurrentTime), metaData); err != nil {
			s.logger.Error("insert device event", zap.Error(err))
		}
	}
	return nil
}

func (s *AcsServer) handleTransferComplete(ctx context.Context, device Device, req *proto.TransferComplete) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	if req == nil {
		return nil
	}
	s.logger.Debug("TransferComplete", zap.Any("CommandKey", req.CommandKey))

	call := device.GetMethodCall(req.CommandKey)
	if call != nil {
		if ts := cast.ToInt64(req.CommandKey); ts != 0 {
			if req.FaultStruct.FaultCode == 0 {
				if err := device.UpdateTransferLogComplete(ts,
					proto.MustParseTime(req.StartTime),
					proto.MustParseTime(req.CompleteTime),
				); err != nil {
					s.logger.Error("update device transfer log with complete", zap.Error(err))
				}
			} else {
				if err := device.UpdateTransferLogFault(ts,
					proto.MustParseTime(req.StartTime),
					proto.MustParseTime(req.CompleteTime),
					req.FaultStruct.FaultCode,
					req.FaultStruct.FaultString,
				); err != nil {
					s.logger.Error("update device transfer log with fault", zap.Error(err))
				}
			}
		}
	}
	return nil
}

func (s *AcsServer) handleAutonomousTransferComplete(ctx context.Context, device Device, req *proto.AutonomousTransferComplete) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	if req == nil {
		return nil
	}

	s.logger.Debug("AutonomousTransferComplete")
	filename := path.Base(req.TransferURL)
	if filename == "." || filename == "/" {
		s.logger.Error("update device transfer log with fault, invalid filename",
			zap.String("filename", filename),
			zap.String("url", req.TransferURL))
		return nil
	}

	fileType, startTime, _, _, _, _, err := pm.ParseFileName(filename)
	if err != nil {
		s.logger.Error("update device transfer log with fault, parse filename", zap.Error(err), zap.String("filename", filename))
		return nil
	}

	key := filename
	ts := startTime.UnixNano()
	if err := device.InsertTransferLogComplete(ts,
		s.uploadBucket,
		key,
		fileType,
		filename,
		proto.MustParseTime(req.StartTime),
		proto.MustParseTime(req.CompleteTime),
		req.FaultStruct.FaultCode,
		req.FaultStruct.FaultString,
	); err != nil {
		s.logger.Error("update device transfer log with fault", zap.Error(err))
	}
	return nil
}

func (s *AcsServer) handleGetRPCMethodsResponse(ctx context.Context, device Device, id string, resp *proto.GetRPCMethodsResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("GetRPCMethodsResponse", zap.Any("MethodList", resp.MethodList.Strings))
	device.UpdateMethodCallResponse(id, jsontype.StringArrayToTags(resp.MethodList.Strings), 0, "")

	methods := resp.MethodList.Strings
	if len(methods) == 0 {
		methods = []string{
			"SetParameterValues",
			"GetParameterValues",
			"GetParameterNames",
			"SetParameterAttributes",
			"GetParameterAttributes",
			"AddObject",
			"DeleteObject",
			"Reboot",
			"Download",
			"Upload",
			"FactoryReset",
		}
	}
	if err := device.UpdateMethods(methods); err != nil {
		return errors.Wrap(err, "update methods of device")
	}
	return nil
}

func (s *AcsServer) handleGetParameterValuesResponse(ctx context.Context, device Device, id string, resp *proto.GetParameterValuesResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("GetParameterValuesResponse", zap.Int("ParameterList", len(resp.ParameterList.ParameterValueStructs)))
	values := map[string]string{}
	for _, v := range resp.ParameterList.ParameterValueStructs {
		values[v.Name.Text] = v.Value.Text
	}
	device.UpdateMethodCallResponse(id, jsontype.StringMapToTags(values), 0, "")
	device.UpdateParameterValues(values)

	if product := device.Product(); product != nil {
		if dm := product.DataModel(); dm != nil {
			needReload := false
			if !GetState("sync_datamodel_param_type", fmt.Sprintf("%v", product.ID())) {
				for _, v := range resp.ParameterList.ParameterValueStructs {
					name := v.Name.Text
					var typ *string
					for _, attr := range v.Value.Attrs {
						if attr.Name.Local == "type" {
							typ = &attr.Value
						}
					}
					dm.UpsertParameter(name, typ, nil, nil, nil)
					needReload = true
				}
			}
			if needReload {
				SetState("sync_datamodel_param_type", fmt.Sprintf("%v", product.ID()), true)
				go dm.Reload()
			}
		}
	}
	return nil
}

func (s *AcsServer) handleSetParameterValuesResponse(ctx context.Context, device Device, id string, resp *proto.SetParameterValuesResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("SetParameterValuesResponse", zap.Int("Status", resp.Status))
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}

func (s *AcsServer) handleGetParameterNamesResponse(ctx context.Context, device Device, id string, resp *proto.GetParameterNamesResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	if resp == nil {
		return nil
	}
	s.logger.Debug("GetParameterNamesResponse", zap.Int("Length", len(resp.ParameterList.ParameterInfoStructs)))

	values := map[string]bool{}
	for _, v := range resp.ParameterList.ParameterInfoStructs {
		values[v.Name] = cast.ToBool(v.Writable)
	}
	device.UpdateMethodCallResponse(id, jsontype.BoolMapToTags(values), 0, "")
	device.UpdateParameterWritables(values)

	if device.IsMethodSupported("GetParameterValues") {
		values := map[string]any{}
		values["Device."] = ""
		device.PushMethodCall(time.Now(), "GetParameterValues", values)
	}

	if product := device.Product(); product != nil {
		if dm := product.DataModel(); dm != nil {
			needReload := false
			if !GetState("sync_datamodel_param_writable", fmt.Sprintf("%v", product.ID())) {
				for _, v := range resp.ParameterList.ParameterInfoStructs {
					name := v.Name
					var writable *bool
					*writable = cast.ToBool(v.Writable)
					dm.UpsertParameter(name, nil, writable, nil, nil)
					needReload = true
				}
			}
			if needReload {
				SetState("sync_datamodel_param_writable", fmt.Sprintf("%v", product.ID()), true)
				go dm.Reload()
			}
		}
	}

	return nil
}

func (s *AcsServer) handleGetParameterAttributesResponse(ctx context.Context, device Device, id string, resp *proto.GetParameterAttributesResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	if resp == nil {
		return nil
	}
	s.logger.Debug("GetParameterAttributesResponse", zap.Int("Length", len(resp.ParameterList.ParameterAttributesStructs)))
	values := map[string]int{}
	for _, v := range resp.ParameterList.ParameterAttributesStructs {
		values[v.Name] = v.Notification
	}
	device.UpdateMethodCallResponse(id, jsontype.IntMapToTags(values), 0, "")
	device.UpdateParameterNotifications(values)
	return nil
}

func (s *AcsServer) handleSetParameterAttributesResponse(ctx context.Context, device Device, id string, resp *proto.SetParameterAttributesResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("SetParameterAttributesResponse")
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}

func (s *AcsServer) handleAddObjectResponse(ctx context.Context, device Device, id string, resp *proto.AddObjectResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("AddObjectResponse",
		zap.Uint("InstanceNumber", resp.InstanceNumber),
		zap.Int("Status", resp.Status),
	)
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}

func (s *AcsServer) handleDeleteObjectResponse(ctx context.Context, device Device, id string, resp *proto.DeleteObjectResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("DeleteObjectResponse",
		zap.Int("Status", resp.Status),
	)
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}

func (s *AcsServer) handleDownloadResponse(ctx context.Context, device Device, id string, resp *proto.DownloadResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("DownloadResponse",
		zap.Int("Status", resp.Status),
		zap.String("StartTime", resp.StartTime),
		zap.String("CompleteTime", resp.CompleteTime),
	)
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}

func (s *AcsServer) handleUploadResponse(ctx context.Context, device Device, id string, resp *proto.UploadResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("UploadResponse",
		zap.Int("Status", resp.Status),
		zap.String("StartTime", resp.StartTime),
		zap.String("CompleteTime", resp.CompleteTime),
	)
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}

func (s *AcsServer) handleRebootResponse(ctx context.Context, device Device, id string, resp *proto.RebootResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("RebootResponse")
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}

func (s *AcsServer) handleFactoryResetResponse(ctx context.Context, device Device, id string, resp *proto.FactoryResetResponse) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	if resp == nil {
		return nil
	}
	s.logger.Debug("FactoryResetResponse")
	device.UpdateMethodCallResponse(id, nil, 0, "")
	return nil
}
