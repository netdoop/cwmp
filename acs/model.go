package acs

import (
	"time"

	"github.com/netdoop/cwmp/proto"
)

type DataModel interface {
	Reload()
	UpsertParameter(name string, typ *string, writable *bool, description *string, defaultValue *string) error
	GetParameterType(name string) string
}

type Product interface {
	GetID() uint
	GetSchema() string
	GetDataModel() DataModel
}

type Device interface {
	GetID() uint
	GetSchema() string
	GetOUI() string
	GetProductClass() string
	GetSerialNumber() string
	GetProduct() Product

	SaveUpdated(updateColumns []string) error

	HandleInform(inform *proto.Inform) error
	GetOnlineStatus() bool
	HandleAlive(t time.Time, lastOnlineStatus bool)

	UpdateMethods(methods []string) error
	IsMethodSupported(method string) bool

	GetNextMethodCall() MethodCall
	GetMethodCall(commandKey string) MethodCall
	PushMethodCall(t time.Time, methodName string, values map[string]any) (MethodCall, error)
	UpdateMethodCallRequestSend(commandKey string) error
	UpdateMethodCallResponse(requestId string, values map[string]any, faultÇode int, faultString string)
	UpdateMethodCallUnknow(commandKey string) error

	UpdateTransferLogComplete(ts int64, startTime time.Time, completeTime time.Time) error
	UpdateTransferLogFault(ts int64, startTime time.Time, completeTime time.Time, faultCode int, faultString string) error
	InsertTransferLogComplete(ts int64, bucket string, key string, fileType string, fileName string, startTime time.Time, completeTime time.Time, faultCode int, faultString string) error

	UpdateParameterWritables(values map[string]bool)
	UpdateParameterNotifications(values map[string]int)

	SyncDeviceParameterNames(parameterPath string, nextLevel bool) error

	GetParameterNames(path string, nextLevel bool) []string
	GetParameterValue(name string) (string, error)
	MustGetParameterValue(name string) string
	UpdateParameterValues(values map[string]string) error

	InsertEvent(eventType string, currentTime time.Time, metaData map[string]any) error
}

type MethodCall interface {
	GetMethodName() string
	GetCommandKey() string
	GetRequestValues() map[string]string
	GetRequestValue(name string) string
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
