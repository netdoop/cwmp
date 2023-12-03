package proto

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

const (
	XMLNS_XSD      = "http://www.w3.org/2001/XMLSchema"
	XMLNS_XSI      = "http://www.w3.org/2001/XMLSchema-instance"
	XMLNS_SOAP_ENC = "http://schemas.xmlsoap.org/soap/encoding/"
	XMLNS_SOAP_ENV = "http://schemas.xmlsoap.org/soap/envelope/"

	XMLNS_CWMP_1_0 = "urn:dslforum-org:cwmp-1-0"
	XMLNS_CWMP_1_1 = "urn:dslforum-org:cwmp-1-1"
	XMLNS_CWMP_1_2 = "urn:dslforum-org:cwmp-1-2"
	XMLNS_CWMP_1_3 = "urn:dslforum-org:cwmp-1-3"
	XMLNS_CWMP_1_4 = "urn:dslforum-org:cwmp-1-4"
)

type SoapNamespace struct {
	SoapEnv xml.Attr `xml:"-"`
	SoapEnc xml.Attr `xml:"-"`
	Cwmp    xml.Attr `xml:"-"`
	Xsd     xml.Attr `xml:"-"`
	Xsi     xml.Attr `xml:"-"`
}

func (m SoapNamespace) ToString() string {
	v, _ := json.Marshal(m)
	return string(v)
}
func (m *SoapNamespace) FromString(in string) error {
	return json.Unmarshal([]byte(in), m)
}

type SoapEnvelope struct {
	XMLName xml.Name
	Text    string        `xml:",chardata"`
	Attrs   []xml.Attr    `xml:",any,attr"`
	NS      SoapNamespace `xml:"-"`

	Header SoapHeader `xml:"Header"`
	Body   SoapBody   `xml:"Body"`
}

func (m *SoapEnvelope) GetNamespace() SoapNamespace {
	ns := SoapNamespace{}
	for _, attr := range m.Attrs {
		switch attr.Value {
		case XMLNS_SOAP_ENC:
			ns.SoapEnc = attr
		case XMLNS_SOAP_ENV:
			ns.SoapEnv = attr
		case XMLNS_CWMP_1_0:
			ns.Cwmp = attr
		case XMLNS_CWMP_1_1:
			ns.Cwmp = attr
		case XMLNS_CWMP_1_2:
			ns.Cwmp = attr
		case XMLNS_CWMP_1_3:
			ns.Cwmp = attr
		case XMLNS_CWMP_1_4:
			ns.Cwmp = attr
		case XMLNS_XSD:
			ns.Xsd = attr
		case XMLNS_XSI:
			ns.Xsi = attr
		}
	}
	return ns
}

func (m *SoapEnvelope) Decode() error {
	m.NS = m.GetNamespace()
	return nil
}

func (m *SoapEnvelope) Encode() error {
	ns := m.NS
	m.Attrs = []xml.Attr{
		makeXmlAttr(ns.SoapEnv),
		makeXmlAttr(ns.SoapEnc),
		makeXmlAttr(ns.Xsd),
		makeXmlAttr(ns.Xsi),
		makeXmlAttr(ns.Cwmp),
	}
	m.XMLName = makeXmlName(ns.SoapEnv, "Envelope")
	m.Header.XMLName = makeXmlName(ns.SoapEnv, "Header")
	if m.Header.ID != nil {
		m.Header.ID.XMLName = makeXmlName(ns.Cwmp, "ID")
		m.Header.ID.Attrs = makeXmlAttrs(ns.SoapEnv, m.Header.ID.Attrs)
	}
	m.Body.XMLName = makeXmlName(ns.SoapEnv, "Body")
	if m.Body.Fault != nil {
		m.Body.Fault.XMLName = makeXmlName(ns.SoapEnv, "Fault")
		m.Body.Fault.Detail.Fault.XMLName = makeXmlName(ns.Cwmp, "Fault")
	}

	if m.Body.Inform != nil {
		m.Body.Inform.XMLName = makeXmlName(ns.Cwmp, "Inform")
	}
	if m.Body.InformResponse != nil {
		m.Body.InformResponse.XMLName = makeXmlName(ns.Cwmp, "InformResponse")
	}
	if m.Body.TransferComplete != nil {
		m.Body.TransferComplete.XMLName = makeXmlName(ns.Cwmp, "TransferComplete")
	}
	if m.Body.TransferCompleteResponse != nil {
		m.Body.TransferCompleteResponse.XMLName = makeXmlName(ns.Cwmp, "TransferCompleteResponse")
	}
	if m.Body.AutonomousTransferComplete != nil {
		m.Body.AutonomousTransferComplete.XMLName = makeXmlName(ns.Cwmp, "AutonomousTransferComplete")
	}
	if m.Body.AutonomousTransferCompleteResponse != nil {
		m.Body.AutonomousTransferCompleteResponse.XMLName = makeXmlName(ns.Cwmp, "AutonomousTransferCompleteResponse")
	}

	if m.Body.GetRPCMethods != nil {
		m.Body.GetRPCMethods.XMLName = makeXmlName(ns.Cwmp, "GetRPCMethods")
	}
	if m.Body.GetRPCMethodsResponse != nil {
		m.Body.GetRPCMethodsResponse.XMLName = makeXmlName(ns.Cwmp, "GetRPCMethodsResponse")
		m.Body.GetRPCMethodsResponse.MethodList.Attrs = []xml.Attr{
			{
				Name:  makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:string[%v]", ns.Xsd.Name.Local, len(m.Body.GetRPCMethodsResponse.MethodList.Strings)),
			},
		}
	}
	if m.Body.SetParameterValues != nil {
		m.Body.SetParameterValues.XMLName = makeXmlName(ns.Cwmp, "SetParameterValues")
		m.Body.SetParameterValues.ParameterList.Attrs = []xml.Attr{
			{
				Name: makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:ParameterValueStruct[%v]",
					ns.Cwmp.Name.Local, len(m.Body.SetParameterValues.ParameterList.ParameterValueStructs)),
			},
		}
		for i := 0; i < len(m.Body.SetParameterValues.ParameterList.ParameterValueStructs); i++ {
			m.Body.SetParameterValues.ParameterList.ParameterValueStructs[i].Name.Attrs = []xml.Attr{
				{
					Name:  makeXmlName(ns.Xsi, "type"),
					Value: fmt.Sprintf("%v:string", ns.Xsd.Name.Local),
				},
			}
			m.Body.SetParameterValues.ParameterList.ParameterValueStructs[i].Value.Attrs = []xml.Attr{
				{
					Name: makeXmlName(ns.Xsi, "type"),
					Value: fmt.Sprintf("%v:%v", ns.Xsd.Name.Local,
						m.Body.SetParameterValues.ParameterList.ParameterValueStructs[i].Value.TypeName,
					),
				},
			}
		}
	}
	if m.Body.SetParameterValuesResponse != nil {
		m.Body.SetParameterValuesResponse.XMLName = makeXmlName(ns.Cwmp, "SetParameterValuesResponse")
	}
	if m.Body.GetParameterValues != nil {
		m.Body.GetParameterValues.XMLName = makeXmlName(ns.Cwmp, "GetParameterValues")
		m.Body.GetParameterValues.ParameterNames.Attrs = []xml.Attr{
			{
				Name:  makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:string[%v]", ns.Xsd.Name.Local, len(m.Body.GetParameterValues.ParameterNames.Strings)),
			},
		}
	}
	if m.Body.GetParameterValuesResponse != nil {
		m.Body.GetParameterValuesResponse.XMLName = makeXmlName(ns.Cwmp, "GetParameterValuesResponse")
		m.Body.GetParameterValuesResponse.ParameterList.Attrs = []xml.Attr{
			{
				Name: makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:ParameterValueStruct[%v]", ns.Cwmp.Name.Local,
					len(m.Body.GetParameterValuesResponse.ParameterList.ParameterValueStructs)),
			},
		}
	}
	if m.Body.GetParameterNames != nil {
		m.Body.GetParameterNames.XMLName = makeXmlName(ns.Cwmp, "GetParameterNames")
	}
	if m.Body.GetParameterNamesResponse != nil {
		m.Body.GetParameterNamesResponse.XMLName = makeXmlName(ns.Cwmp, "GetParameterNamesResponse")
		m.Body.GetParameterNamesResponse.ParameterList.Attrs = []xml.Attr{
			{
				Name: makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:ParameterInfoStruct[%v]", ns.Cwmp.Name.Local,
					len(m.Body.GetParameterNamesResponse.ParameterList.ParameterInfoStructs)),
			},
		}
	}
	if m.Body.SetParameterAttributes != nil {
		m.Body.SetParameterAttributes.XMLName = makeXmlName(ns.Cwmp, "SetParameterAttributes")
		m.Body.SetParameterAttributes.ParameterList.Attrs = []xml.Attr{
			{
				Name: makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:SetParameterAttributesStruct[%v]", ns.Cwmp.Name.Local,
					len(m.Body.SetParameterAttributes.ParameterList.SetParameterAttributesStructs)),
			},
		}
		for _, v := range m.Body.SetParameterAttributes.ParameterList.SetParameterAttributesStructs {
			v.AccessList.Attrs = []xml.Attr{
				{
					Name: makeXmlName(ns.SoapEnc, "arrayType"),
					Value: fmt.Sprintf("%v:string[%v]", ns.Xsd.Name.Local,
						len(v.AccessList.Strings)),
				},
			}
		}
	}
	if m.Body.SetParameterAttributesResponse != nil {
		m.Body.SetParameterAttributesResponse.XMLName = makeXmlName(ns.Cwmp, "SetParameterAttributesResponse")
	}
	if m.Body.GetParameterAttributes != nil {
		m.Body.GetParameterAttributes.XMLName = makeXmlName(ns.Cwmp, "GetParameterAttributes")
		m.Body.GetParameterAttributes.ParameterNames.Attrs = []xml.Attr{
			{
				Name: makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:string[%v]", ns.Xsd.Name.Local,
					len(m.Body.GetParameterAttributes.ParameterNames.Strings)),
			},
		}
	}
	if m.Body.GetParameterAttributesResponse != nil {
		m.Body.GetParameterAttributesResponse.XMLName = makeXmlName(ns.Cwmp, "GetParameterAttributesResponse")
		m.Body.GetParameterAttributesResponse.ParameterList.Attrs = []xml.Attr{
			{
				Name: makeXmlName(ns.SoapEnc, "arrayType"),
				Value: fmt.Sprintf("%v:ParameterAttributesStruct[%v]", ns.Cwmp.Name.Local,
					len(m.Body.GetParameterAttributesResponse.ParameterList.ParameterAttributesStructs)),
			},
		}
		for _, v := range m.Body.GetParameterAttributesResponse.ParameterList.ParameterAttributesStructs {
			v.AccessList.Attrs = []xml.Attr{
				{
					Name: makeXmlName(ns.SoapEnc, "arrayType"),
					Value: fmt.Sprintf("%v:string[%v]", ns.Xsd.Name.Local,
						len(v.AccessList.Strings)),
				},
			}
		}
	}
	if m.Body.AddObject != nil {
		m.Body.AddObject.XMLName = makeXmlName(ns.Cwmp, "AddObject")
	}
	if m.Body.AddObjectResponse != nil {
		m.Body.AddObjectResponse.XMLName = makeXmlName(ns.Cwmp, "AddObjectResponse")
	}
	if m.Body.DeleteObject != nil {
		m.Body.DeleteObject.XMLName = makeXmlName(ns.Cwmp, "DeleteObject")
	}
	if m.Body.DeleteObjectResponse != nil {
		m.Body.DeleteObjectResponse.XMLName = makeXmlName(ns.Cwmp, "DeleteObjectResponse")
	}
	if m.Body.Download != nil {
		m.Body.Download.XMLName = makeXmlName(ns.Cwmp, "Download")
	}
	if m.Body.DownloadResponse != nil {
		m.Body.DownloadResponse.XMLName = makeXmlName(ns.Cwmp, "DownloadResponse")
	}
	if m.Body.Upload != nil {
		m.Body.Upload.XMLName = makeXmlName(ns.Cwmp, "Upload")
	}
	if m.Body.UploadResponse != nil {
		m.Body.UploadResponse.XMLName = makeXmlName(ns.Cwmp, "UploadResponse")
	}
	if m.Body.Reboot != nil {
		m.Body.Reboot.XMLName = makeXmlName(ns.Cwmp, "Reboot")
	}
	if m.Body.RebootResponse != nil {
		m.Body.RebootResponse.XMLName = makeXmlName(ns.Cwmp, "RebootResponse")
	}
	if m.Body.FactoryReset != nil {
		m.Body.FactoryReset.XMLName = makeXmlName(ns.Cwmp, "FactoryReset")
	}
	if m.Body.FactoryResetResponse != nil {
		m.Body.FactoryResetResponse.XMLName = makeXmlName(ns.Cwmp, "FactoryResetResponse")
	}

	return nil
}

type CwmpID struct {
	XMLName xml.Name
	Text    string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",any,attr"`
}
type HoldRequests struct {
	XMLName        xml.Name
	Text           string `xml:",chardata"`
	MustUnderstand string `xml:"mustUnderstand,attr"` // must "1"
}
type SessionTimeout struct {
	XMLName        xml.Name
	Text           string `xml:",chardata"`
	MustUnderstand string `xml:"mustUnderstand,attr"` // must "0"
}
type SupportedCWMPVersions struct {
	XMLName        xml.Name
	Text           string `xml:",chardata"`
	MustUnderstand string `xml:"mustUnderstand,attr"` // must "0"
}
type UseCWMPVersion struct {
	XMLName        xml.Name
	Text           string `xml:",chardata"`
	MustUnderstand string `xml:"mustUnderstand,attr"` // must "1"
}

type SoapHeader struct {
	XMLName               xml.Name
	Text                  string                 `xml:",chardata"`
	ID                    *CwmpID                `xml:"ID,omitempty"`
	HoldRequests          *HoldRequests          `xml:"HoldRequests,omitempty"`          // ACS to CPE
	SessionTimeout        *SessionTimeout        `xml:"SessionTimeout,omitempty"`        // CPE to ACS
	SupportedCWMPVersions *SupportedCWMPVersions `xml:"SupportedCWMPVersions,omitempty"` // CPE to ACS
	UseCWMPVersion        *UseCWMPVersion        `xml:"UseCWMPVersion,omitempty"`        // ACS to CPE
}

type SoapBody struct {
	XMLName xml.Name
	Text    string     `xml:",chardata"`
	Fault   *SoapFault `xml:"Fault,omitempty"`

	GetRPCMethods                  *GetRPCMethods                  `xml:"GetRPCMethods,omitempty"`
	GetRPCMethodsResponse          *GetRPCMethodsResponse          `xml:"GetRPCMethodsResponse,omitempty"`
	SetParameterValues             *SetParameterValues             `xml:"SetParameterValues,omitempty"`
	SetParameterValuesResponse     *SetParameterValuesResponse     `xml:"SetParameterValuesResponse,omitempty"`
	GetParameterValues             *GetParameterValues             `xml:"GetParameterValues,omitempty"`
	GetParameterValuesResponse     *GetParameterValuesResponse     `xml:"GetParameterValuesResponse,omitempty"`
	GetParameterNames              *GetParameterNames              `xml:"GetParameterNames,omitempty"`
	GetParameterNamesResponse      *GetParameterNamesResponse      `xml:"GetParameterNamesResponse,omitempty"`
	GetParameterAttributes         *GetParameterAttributes         `xml:"GetParameterAttributes,omitempty"`
	GetParameterAttributesResponse *GetParameterAttributesResponse `xml:"GetParameterAttributesResponse,omitempty"`
	SetParameterAttributes         *SetParameterAttributes         `xml:"SetParameterAttributes,omitempty"`
	SetParameterAttributesResponse *SetParameterAttributesResponse `xml:"SetParameterAttributesResponse,omitempty"`
	AddObject                      *AddObject                      `xml:"AddObject,omitempty"`
	AddObjectResponse              *AddObjectResponse              `xml:"AddObjectResponse,omitempty"`
	DeleteObject                   *DeleteObject                   `xml:"DeleteObject,omitempty"`
	DeleteObjectResponse           *DeleteObjectResponse           `xml:"DeleteObjectResponse,omitempty"`
	Download                       *Download                       `xml:"Download,omitempty"`
	DownloadResponse               *DownloadResponse               `xml:"DownloadResponse,omitempty"`
	Upload                         *Upload                         `xml:"Upload,omitempty"`
	UploadResponse                 *UploadResponse                 `xml:"UploadResponse,omitempty"`
	Reboot                         *Reboot                         `xml:"Reboot,omitempty"`
	RebootResponse                 *RebootResponse                 `xml:"RebootResponse,omitempty"`
	FactoryReset                   *FactoryReset                   `xml:"FactoryReset,omitempty"`
	FactoryResetResponse           *FactoryResetResponse           `xml:"FactoryResetResponse,omitempty"`

	Inform                             *Inform                             `xml:"Inform,omitempty"`
	InformResponse                     *InformResponse                     `xml:"InformResponse,omitempty"`
	TransferComplete                   *TransferComplete                   `xml:"TransferComplete,omitempty"`
	TransferCompleteResponse           *TransferCompleteResponse           `xml:"TransferCompleteResponse,omitempty"`
	AutonomousTransferComplete         *AutonomousTransferComplete         `xml:"AutonomousTransferComplete,omitempty"`
	AutonomousTransferCompleteResponse *AutonomousTransferCompleteResponse `xml:"AutonomousTransferCompleteResponse,omitempty"`
}

type SoapFault struct {
	XMLName     xml.Name
	Text        string          `xml:",chardata"`
	FaultCode   string          `xml:"faultcode"`
	FaultString string          `xml:"faultstring"`
	Detail      SoapFaultDetail `xml:"detail"`
}

func CreateSoapFault(code int, err error) *SoapFault {
	fault := &SoapFault{
		FaultCode: fmt.Sprintf("%v", code),
	}
	if err != nil {
		fault.FaultString = err.Error()
	}
	return fault
}

func (m SoapFault) String() string {
	return fmt.Sprintf("[Code=%v String=%v Deailt=%v]", m.FaultCode, m.FaultString, m.Detail)
}

type SoapFaultDetail struct {
	XMLName xml.Name
	Text    string    `xml:",chardata"`
	Fault   CwmpFault `xml:"Fault"`
}

func (m SoapFaultDetail) String() string {
	return fmt.Sprintf("%v", m.Fault)
}

type CwmpFault struct {
	XMLName                  xml.Name
	Text                     string                    `xml:",chardata"`
	FaultCode                string                    `xml:"FaultCode"`
	FaultString              string                    `xml:"FaultString"`
	SetParameterValuesFaults []SetParameterValuesFault `xml:"SetParameterValuesFault"`
}

func (m CwmpFault) String() string {
	return fmt.Sprintf("[Code=%v String=%v]", m.FaultCode, m.FaultString)
}

type SetParameterValuesFault struct {
	XMLName       xml.Name
	Text          string `xml:",chardata"`
	ParameterName string `xml:"ParameterName"`
	FaultCode     int    `xml:"FaultCode"`
	FaultString   string `xml:"FaultString"`
}

type DeviceID struct {
	XMLName      xml.Name
	Text         string `xml:",chardata"`
	Manufacturer string `xml:"Manufacturer"`
	OUI          string `xml:"OUI"`
	ProductClass string `xml:"ProductClass"`
	SerialNumber string `xml:"SerialNumber"`
}

func (m DeviceID) String() string {
	return fmt.Sprintf("[MF=%v OUI=%v PC=%v SN=%v]", m.Manufacturer, m.OUI, m.ProductClass, m.SerialNumber)
}

type EventStruct struct {
	XMLName    xml.Name
	Text       string `xml:",chardata"`
	EventCode  string `xml:"EventCode"`
	CommandKey string `xml:"CommandKey"`
}

func (m EventStruct) String() string {
	return fmt.Sprintf("[Code=%v CommandKey=%v]", m.EventCode, m.CommandKey)
}

type Event struct {
	XMLName xml.Name
	Text    string        `xml:",chardata"`
	Events  []EventStruct `xml:"EventStruct"`
}

func (m Event) String() string {
	out := []string{}
	for _, v := range m.Events {
		out = append(out, v.String())
	}
	return fmt.Sprintf("[%v]", strings.Join(out, ", "))
}

type FaultStruct struct {
	XMLName     xml.Name
	Text        string `xml:",chardata"`
	FaultCode   int    `xml:"FaultCode"`
	FaultString string `xml:"FaultString"`
}

type MethodList struct {
	XMLName xml.Name
	Text    string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",any,attr"`
	Strings []string   `xml:"string"`
	//ArrayType string   `xml:"soap-enc:arrayType,attr"`
}

type ParameterNames struct {
	XMLName xml.Name
	Text    string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",any,attr"`
	Strings []string   `xml:"string"`
	//ArrayType string   `xml:"soap-enc:arrayType,attr"`
}

type AccessList struct {
	XMLName xml.Name
	Text    string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",any,attr"`
	Strings []string   `xml:"string"`
	//ArrayType string   `xml:"soap-enc:arrayType,attr"`
}

type ParameterValueStruct struct {
	XMLName xml.Name
	Text    string         `xml:",chardata"`
	Name    ParameterName  `xml:"Name"`
	Value   ParameterValue `xml:"Value"`
}

type ParameterName struct {
	XMLName xml.Name
	Text    string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",any,attr"`
}
type ParameterValue struct {
	XMLName  xml.Name
	Text     string     `xml:",chardata"`
	Attrs    []xml.Attr `xml:",any,attr"`
	TypeName string     `xml:"-"`
}

func (m ParameterValueStruct) String() string {
	return fmt.Sprintf("%v=%v", m.Name, m.Value)
}

type Value struct {
	XMLName xml.Name
	Text    string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",any,attr"`
	//Type    string     `xml:"xsi:type,attr"`
}

func (m Value) String() string {
	return fmt.Sprintf("%v", m.Text)
}

type ParameterInfoStruct struct {
	XMLName  xml.Name
	Text     string `xml:",chardata"`
	Name     string `xml:"Name"`
	Writable string `xml:"Writable"`
}

type SetParameterAttributesStruct struct {
	XMLName            xml.Name
	Text               string     `xml:",chardata"`
	Name               string     `xml:"Name"`
	NotificationChange bool       `xml:"NotificationChange,omitempty"`
	Notification       int        `xml:"Notification,omitempty"`
	AccessListChange   bool       `xml:"AccessListChange,omitempty"`
	AccessList         AccessList `xml:"AccessList,omitempty"`
}
type ParameterAttributesStruct struct {
	XMLName      xml.Name
	Text         string     `xml:",chardata"`
	Name         string     `xml:"Name"`
	Notification int        `xml:"Notification"`
	AccessList   AccessList `xml:"AccessList"`
}

type ParameterListOfParameterValue struct {
	XMLName               xml.Name
	Text                  string                  `xml:",chardata"`
	Attrs                 []xml.Attr              `xml:",any,attr"`
	ParameterValueStructs []*ParameterValueStruct `xml:"ParameterValueStruct"`
	//ArrayType             string                 `xml:"soap-enc:arrayType,attr"`
}

func (m ParameterListOfParameterValue) String() string {
	out := []string{}
	for _, v := range m.ParameterValueStructs {
		out = append(out, v.String())
	}
	return strings.Join(out, "\n")
}

type ParameterListOfParameterInfo struct {
	XMLName              xml.Name
	Text                 string                 `xml:",chardata"`
	Attrs                []xml.Attr             `xml:",any,attr"`
	ParameterInfoStructs []*ParameterInfoStruct `xml:"ParameterInfoStruct"`
	//ArrayType            string                `xml:"soap-enc:arrayType,attr"`
}

func (m ParameterListOfParameterInfo) String() string {
	out := []string{}
	for _, v := range m.ParameterInfoStructs {
		out = append(out, fmt.Sprintf("%v=%v", v.Name, v.Writable))
	}
	return strings.Join(out, "\n")
}

type ParameterListOfSetParameterAttribute struct {
	XMLName                       xml.Name
	Text                          string                          `xml:",chardata"`
	Attrs                         []xml.Attr                      `xml:",any,attr"`
	SetParameterAttributesStructs []*SetParameterAttributesStruct `xml:"SetParameterAttributesStruct"`
	//ArrayType                     string                         `xml:"soap-enc:arrayType,attr"`
}
type ParameterListOfParameterAttribute struct {
	XMLName                    xml.Name
	Text                       string                       `xml:",chardata"`
	Attrs                      []xml.Attr                   `xml:",any,attr"`
	ParameterAttributesStructs []*ParameterAttributesStruct `xml:"ParameterAttributesStruct"`
	//ArrayType                  string                     `xml:"soap-enc:arrayType,attr"`
}

type Inform struct {
	XMLName       xml.Name
	Text          string                        `xml:",chardata"`
	DeviceID      DeviceID                      `xml:"DeviceId"`
	Event         Event                         `xml:"Event"`
	MaxEnvelopes  int                           `xml:"MaxEnvelopes"`
	CurrentTime   string                        `xml:"CurrentTime"`
	RetryCount    int                           `xml:"RetryCount"`
	ParameterList ParameterListOfParameterValue `xml:"ParameterList"`
}

func (m Inform) String() string {
	return fmt.Sprintf("[DeviceID=%v CurrentTime=%v MaxEnvelopes=%v RetryCount=%v Event=%v]",
		m.DeviceID, m.CurrentTime, m.MaxEnvelopes, m.RetryCount, m.Event,
	)
}

type InformResponse struct {
	XMLName      xml.Name
	Text         string `xml:",chardata"`
	MaxEnvelopes int    `xml:"MaxEnvelopes"`
}

type TransferComplete struct {
	XMLName      xml.Name
	Text         string      `xml:",chardata"`
	CommandKey   string      `xml:"CommandKey"`
	FaultStruct  FaultStruct `xml:"FaultStruct"`
	StartTime    string      `xml:"StartTime"`
	CompleteTime string      `xml:"CompleteTime"`
}
type TransferCompleteResponse struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

type AutonomousTransferComplete struct {
	XMLName        xml.Name
	Text           string      `xml:",chardata"`
	AnnounceURL    string      `xml:"AnnounceURL"`
	TransferURL    string      `xml:"TransferURL"`
	IsDownload     bool        `xml:"IsDownload"`
	FileType       string      `xml:"FileType"`
	FileSize       uint        `xml:"FileSize"`
	TargetFileName string      `xml:"TargetFileName"`
	FaultStruct    FaultStruct `xml:"FaultStruct"`
	StartTime      string      `xml:"StartTime"`
	CompleteTime   string      `xml:"CompleteTime"`
}
type AutonomousTransferCompleteResponse struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

type GetRPCMethods struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}
type GetRPCMethodsResponse struct {
	XMLName    xml.Name
	Text       string     `xml:",chardata"`
	MethodList MethodList `xml:"MethodList"`
}

type SetParameterValues struct {
	XMLName       xml.Name
	Text          string                        `xml:",chardata"`
	ParameterList ParameterListOfParameterValue `xml:"ParameterList"`
	ParameterKey  string                        `xml:"ParameterKey"`
}
type SetParameterValuesResponse struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
	Status  int    `xml:"status"`
}

type GetParameterValues struct {
	XMLName        xml.Name
	Text           string         `xml:",chardata"`
	ParameterNames ParameterNames `xml:"ParameterNames"`
}
type GetParameterValuesResponse struct {
	XMLName       xml.Name
	Text          string                        `xml:",chardata"`
	ParameterList ParameterListOfParameterValue `xml:"ParameterList"`
}

type GetParameterNames struct {
	XMLName       xml.Name
	Text          string `xml:",chardata"`
	ParameterPath string `xml:"ParameterPath"`
	NextLevel     string `xml:"NextLevel"`
}
type GetParameterNamesResponse struct {
	XMLName       xml.Name
	Text          string                       `xml:",chardata"`
	ParameterList ParameterListOfParameterInfo `xml:"ParameterList"`
}

type SetParameterAttributes struct {
	XMLName       xml.Name
	Text          string                               `xml:",chardata"`
	ParameterList ParameterListOfSetParameterAttribute `xml:"ParameterList"`
}
type SetParameterAttributesResponse struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

type GetParameterAttributes struct {
	XMLName        xml.Name
	Text           string         `xml:",chardata"`
	ParameterNames ParameterNames `xml:"ParameterNames"`
}
type GetParameterAttributesResponse struct {
	XMLName       xml.Name
	Text          string                            `xml:",chardata"`
	ParameterList ParameterListOfParameterAttribute `xml:"ParameterList"`
}

type AddObject struct {
	XMLName      xml.Name
	Text         string `xml:",chardata"`
	ObjectName   string `xml:"ObjectName"`
	ParameterKey string `xml:"ParameterKey"`
}
type AddObjectResponse struct {
	XMLName        xml.Name
	Text           string `xml:",chardata"`
	InstanceNumber uint   `xml:"InstanceNumber"`
	Status         int    `xml:"status"`
}

type DeleteObject struct {
	XMLName      xml.Name
	Text         string `xml:",chardata"`
	ObjectName   string `xml:"ObjectName"`
	ParameterKey string `xml:"ParameterKey"`
}
type DeleteObjectResponse struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
	Status  int    `xml:"status"`
}

type Download struct {
	XMLName        xml.Name
	Text           string `xml:",chardata"`
	CommandKey     string `xml:"CommandKey"`
	FileType       string `xml:"FileType"`
	URL            string `xml:"URL"`
	Username       string `xml:"Username"`
	Password       string `xml:"Password"`
	FileSize       uint   `xml:"FileSize"`
	TargetFileName string `xml:"TargetFileName"`
	DelaySeconds   uint   `xml:"DelaySeconds"`
	SuccessURL     string `xml:"SuccessURL"`
	FailureURL     string `xml:"FailureURL"`
}
type DownloadResponse struct {
	XMLName      xml.Name
	Text         string `xml:",chardata"`
	Status       int    `xml:"status"`
	StartTime    string `xml:"StartTime"`
	CompleteTime string `xml:"CompleteTime"`
}

type Upload struct {
	XMLName      xml.Name
	Text         string `xml:",chardata"`
	CommandKey   string `xml:"CommandKey"`
	FileType     string `xml:"FileType"`
	URL          string `xml:"URL"`
	Username     string `xml:"Username"`
	Password     string `xml:"Password"`
	DelaySeconds uint   `xml:"DelaySeconds"`
}
type UploadResponse struct {
	XMLName      xml.Name
	Text         string `xml:",chardata"`
	Status       int    `xml:"status"`
	StartTime    string `xml:"StartTime"`
	CompleteTime string `xml:"CompleteTime"`
}

type Reboot struct {
	XMLName    xml.Name
	Text       string `xml:",chardata"`
	CommandKey string `xml:"CommandKey"`
}
type RebootResponse struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

type FactoryReset struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}
type FactoryResetResponse struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}
