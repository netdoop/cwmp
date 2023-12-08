package acs

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/heypkg/s3"
	"github.com/labstack/echo/v4"
	"github.com/netdoop/cwmp/pm"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *AcsServer) HandleUpload(c echo.Context) error {
	name := c.Param("name")
	logger := zap.L()

	req := c.Request()
	contentType := req.Header.Get("Content-Type")
	var (
		buffer   bytes.Buffer
		filename string
	)

	if contentType == "" || strings.HasPrefix(contentType, "text/plain") {
		filename = name
		defer req.Body.Close()
		if _, err := buffer.ReadFrom(req.Body); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "buffer read error").Error())
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "read Object").Error())
		}
		filename = file.Filename
		f, err := file.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "open Object").Error())
		}
		defer f.Close()
		if _, err := buffer.ReadFrom(f); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "buffer read error").Error())
		}
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content type "+contentType)
	}

	fileType, startTime, endTime, oui, _, serialNumber, err := pm.ParseFileName(filename)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "invalid filename").Error())
	}
	interval := s.dataRetentionPeriod

	t := time.Now().Add(interval * -1)
	if startTime.Before(t) && endTime.Before(t) {
		logger.Warn("ignore upload file", zap.String("filename", filename))
		return c.JSON(http.StatusOK, nil)
	}
	schema := ""
	device := s.handler.GetDevice(schema, oui, "", serialNumber)
	if device == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("invalid device").Error())
	}

	switch fileType {
	case "ConfigurationFile":
		key := name
		src := bytes.NewReader(buffer.Bytes())
		obj, err := s3.PutObject(schema, s.uploadBucket, key, name, src)
		if err != nil || obj == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "put object").Error())
		}
	case "LogFile":
		key := name
		src := bytes.NewReader(buffer.Bytes())
		obj, err := s3.PutObject(schema, s.uploadBucket, key, name, src)
		if err != nil || obj == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "put object").Error())
		}
	case "NrmFile":
		key := name
		src := bytes.NewReader(buffer.Bytes())
		obj, err := s3.PutObject(schema, s.uploadBucket, key, name, src)
		if err != nil || obj == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "put object").Error())
		}
	case "PmFile":
		key := name
		src := bytes.NewReader(buffer.Bytes())
		obj, err := s3.PutObject(schema, s.uploadBucket, key, name, src)
		if err != nil || obj == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "put object").Error())
		}

		go func() {
			src := bytes.NewReader(buffer.Bytes())
			collec := &pm.MeasCollecFile{}
			decoder := xml.NewDecoder(src)
			if err := decoder.Decode(collec); err != nil {
				logger.Error("decode PmFile", zap.Error(err))
				return
			}
			// logger.Warn("debug", zap.Any("collec", collec))

			types := map[int]string{}
			for _, data := range collec.MeasData {
				for _, info := range data.MeasInfo {
					for _, typ := range info.MeasTypes {
						types[typ.P] = typ.Value
					}
				}
			}

			values := map[string]any{}
			for _, data := range collec.MeasData {
				for _, info := range data.MeasInfo {
					for _, r := range info.MeasValue.R {
						if typ, ok := types[r.P]; ok {
							tmp := r.Value
							values[typ] = tmp
						}
					}
				}
			}
			s.handler.HandleMesureValues(device, filename, values)
			// for typ, value := range values {
			// 	if meas := omc.GetKPIMeasure(schema, "enb", typ); meas != nil && meas.Enable {
			// 		if v := cast.ToFloat64(value); v != 0 {
			// 			omc.InsertDeviePerformanceValue(db, device, meas.MeasTypeID, filename, v, nil)
			// 		}
			// 	}
			// }

			// measList := omc.GetKPIMeasuresBySet(schema, "enb", "Customize")
			// for _, meas := range measList {
			// 	if meas.Enable && meas.FormulaExpression != nil {
			// 		if result, err := meas.FormulaExpression.Evaluate(values); err != nil {
			// 			logger.Error("evaluate", zap.String("formula", meas.Formula))
			// 		} else {
			// 			v := fmt.Sprintf("%v", result)
			// 			if v != "NaN" {
			// 				omc.InsertDeviePerformanceValue(db, device, meas.MeasTypeID, filename, cast.ToFloat64(v), nil)
			// 			}
			// 		}
			// 	}
			// }
		}()

	}

	return c.JSON(http.StatusOK, nil)
}
