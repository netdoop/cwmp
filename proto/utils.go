package proto

import (
	"encoding/xml"
	"strings"
	"time"
	"unicode/utf8"
)

func CreateEnvelopeFault(id string, ns SoapNamespace, code int, err error) *SoapEnvelope {
	return CreateEnvelope(id, ns, CreateSoapFault(code, err))
}

func CreateEnvelope(id string, ns SoapNamespace, body any) *SoapEnvelope {
	msg := &SoapEnvelope{
		NS: ns,
	}
	msg.Header.ID = &CwmpID{Text: id}
	msg.Header.ID.Attrs = []xml.Attr{
		{
			Name:  xml.Name{Space: "", Local: "mustUnderstand"},
			Value: "1",
		},
	}

	switch v := body.(type) {
	case *Inform:
		msg.Body.Inform = v
	case *InformResponse:
		msg.Body.InformResponse = v
	case *TransferComplete:
		msg.Body.TransferComplete = v
	case *TransferCompleteResponse:
		msg.Body.TransferCompleteResponse = v
	case *AutonomousTransferComplete:
		msg.Body.AutonomousTransferComplete = v
	case *AutonomousTransferCompleteResponse:
		msg.Body.AutonomousTransferCompleteResponse = v
	case *GetRPCMethods:
		msg.Body.GetRPCMethods = v
	case *GetRPCMethodsResponse:
		msg.Body.GetRPCMethodsResponse = v
	case *GetParameterNames:
		msg.Body.GetParameterNames = v
	case *GetParameterNamesResponse:
		msg.Body.GetParameterNamesResponse = v
	case *GetParameterValues:
		msg.Body.GetParameterValues = v
	case *GetParameterValuesResponse:
		msg.Body.GetParameterValuesResponse = v
	case *SetParameterValues:
		msg.Body.SetParameterValues = v
	case *SetParameterValuesResponse:
		msg.Body.SetParameterValuesResponse = v
	case *GetParameterAttributes:
		msg.Body.GetParameterAttributes = v
	case *GetParameterAttributesResponse:
		msg.Body.GetParameterAttributesResponse = v
	case *SetParameterAttributes:
		msg.Body.SetParameterAttributes = v
	case *SetParameterAttributesResponse:
		msg.Body.SetParameterAttributesResponse = v
	case *AddObject:
		msg.Body.AddObject = v
	case *AddObjectResponse:
		msg.Body.AddObjectResponse = v
	case *DeleteObject:
		msg.Body.DeleteObject = v
	case *DeleteObjectResponse:
		msg.Body.DeleteObjectResponse = v
	case *Download:
		msg.Body.Download = v
	case *DownloadResponse:
		msg.Body.DownloadResponse = v
	case *Upload:
		msg.Body.Upload = v
	case *UploadResponse:
		msg.Body.UploadResponse = v
	case *Reboot:
		msg.Body.Reboot = v
	case *RebootResponse:
		msg.Body.RebootResponse = v
	case *FactoryReset:
		msg.Body.FactoryReset = v
	case *FactoryResetResponse:
		msg.Body.FactoryResetResponse = v
	case *SoapFault:
		msg.Body.Fault = v
	}
	msg.Encode()
	return msg
}

func CreateEnvelopeWithRequest(req *SoapEnvelope) *SoapEnvelope {
	resp := &SoapEnvelope{
		NS: req.NS,
		Header: SoapHeader{
			XMLName:      req.Header.XMLName,
			ID:           req.Header.ID,
			HoldRequests: nil, // 0
		},
		Body: SoapBody{
			XMLName: req.Body.XMLName,
		},
	}
	return resp
}

func CleanXMLData(data []byte) []byte {
	var cleanedData strings.Builder
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r != utf8.RuneError || size > 1 {
			if r == '\u0010' {
				cleanedData.WriteRune(' ') // replace illegal character with space
			} else {
				cleanedData.WriteRune(r)
			}
		}
		data = data[size:]
	}
	return []byte(cleanedData.String())
}

func makeXmlName(attr xml.Attr, local string) xml.Name {
	return xml.Name{
		Space: "",
		Local: attr.Name.Local + ":" + local,
	}
}
func makeXmlAttr(src xml.Attr) xml.Attr {
	return xml.Attr{
		Name: xml.Name{
			Space: "",
			Local: src.Name.Space + ":" + src.Name.Local,
		},
		Value: src.Value,
	}
}
func makeXmlAttrs(ns xml.Attr, src []xml.Attr) []xml.Attr {
	out := []xml.Attr{}
	for _, attr := range src {
		out = append(out, xml.Attr{Name: makeXmlName(ns, attr.Name.Local), Value: attr.Value})
	}
	return out
}

func ParseTime(value string) (time.Time, error) {
	var layouts = []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05 MST",
	}
	var out time.Time
	var err error
	for _, layout := range layouts {
		out, err = time.Parse(layout, value)
		if err == nil {
			break
		}
	}
	return out, err
}

func MustParseTime(value string) time.Time {
	t, _ := ParseTime(value)
	return t
}

func FormateTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}
