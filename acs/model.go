package acs

import (
	"time"

	"github.com/netdoop/cwmp/proto"
)

type DataModel interface {
	Reload()
	UpsertParameter(name string, typ *string, writable *bool, description *string, defaultValue *string) error
}

type Product interface {
	ID() string
	DataModel() DataModel
}

type Device interface {
	ID() string
	OUI() string
	ProductClass() string
	SerialNumber() string
	Product() Product

	OnlineStatus() bool
	HandleAlive(t time.Time, lastOnlineStatus bool)

	UpdateMethods(methods []string) error
	IsMethodSupported(method string) bool

	GetMethodCall(commandKey string) MethodCall
	PushMethodCall(t time.Time, methodName string, values map[string]any) (MethodCall, error)
	UpdateMethodCallResponse(requestId string, values map[string]any, faultÇode int, faultString string)

	UpdateTransferLogComplete(ts int64, startTime time.Time, completeTime time.Time) error
	UpdateTransferLogFault(ts int64, startTime time.Time, completeTime time.Time, faultCode int, faultString string) error
	InsertTransferLogComplete(ts int64, bucket string, key string, fileType string, fileName string, startTime time.Time, completeTime time.Time, faultCode int, faultString string) error

	UpdateParameterWritables(values map[string]bool)
	UpdateParameterNotifications(values map[string]int)

	GetDeviceParameterNames(parameterPath string, nextLevel bool) error
	GetParameterValue(name string) (string, error)
	MustGetParameterValue(name string) string
	UpdateParameterValues(values map[string]string) error

	InsertEvent(eventType string, currentTime time.Time, metaData map[string]any) error
}

type MethodCall interface {
}

type ProductStore interface {
	GetProduct(schema string, oui string, productClass string) Product
}

type DeviceStore interface {
	GetDevice(schema string, oui string, productClass string, serialNumber string) Device
	CreateDeviceWithInform(schema string, oui string, productClass string, serialNumber string, inform *proto.Inform) Device
}

type PerformanceHanlder interface {
	HandleMesureValues(values map[string]any)
}
