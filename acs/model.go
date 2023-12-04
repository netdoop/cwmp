package acs

import (
	"context"
	"time"

	"github.com/netdoop/cwmp/proto"
)

// type DataModel interface {
// 	Reload()
// 	UpsertParameter(name string, typ *string, writable *bool, description *string, defaultValue *string) error
// 	GetParameterType(name string) string
// }

// type Product interface {
// 	GetID() uint
// 	GetSchema() string
// 	GetDataModel() DataModel
// }

// type Device interface {
// 	GetID() uint
// 	GetSchema() string
// 	GetOUI() string
// 	GetProductClass() string
// 	GetSerialNumber() string
// 	GetProduct() Product

// 	SaveUpdated(updateColumns []string) error

// 	HandleInform(inform *proto.Inform) error
// 	GetOnlineStatus() bool
// 	HandleAlive(t time.Time, lastOnlineStatus bool)

// 	UpdateMethods(methods []string) error
// 	IsMethodSupported(method string) bool

// 	GetNextMethodCall() MethodCall
// 	GetMethodCall(commandKey string) MethodCall
// 	PushMethodCall(t time.Time, methodName string, values map[string]any) (MethodCall, error)
// 	UpdateMethodCallRequestSend(commandKey string) error
// 	UpdateMethodCallResponse(requestId string, values map[string]any, faultÇode int, faultString string)
// 	UpdateMethodCallUnknow(commandKey string) error

// 	UpdateTransferLogComplete(ts int64, startTime time.Time, completeTime time.Time) error
// 	UpdateTransferLogFault(ts int64, startTime time.Time, completeTime time.Time, faultCode int, faultString string) error
// 	InsertTransferLogComplete(ts int64, bucket string, key string, fileType string, fileName string, startTime time.Time, completeTime time.Time, faultCode int, faultString string) error

// 	UpdateParameterWritables(values map[string]bool)
// 	UpdateParameterNotifications(values map[string]int)

// 	SyncDeviceParameterNames(parameterPath string, nextLevel bool) error

// 	GetParameterNames(path string, nextLevel bool) []string
// 	GetParameterValue(name string) (string, error)
// 	MustGetParameterValue(name string) string
// 	UpdateParameterValues(values map[string]string) error

// 	InsertEvent(eventType string, currentTime time.Time, metaData map[string]any) error
// }

// type MethodCall interface {
// 	GetMethodName() string
// 	GetCommandKey() string
// 	GetRequestValues() map[string]string
// 	GetRequestValue(name string) string
// }

// type ProductStore interface {
// 	GetProduct(schema string, oui string, productClass string) Product
// }

// type DeviceStore interface {
// 	GetDevice(schema string, oui string, productClass string, serialNumber string) Device
// 	CreateDeviceWithInform(schema string, oui string, productClass string, serialNumber string, inform *proto.Inform) Device
// }

type DataModel interface {
	GetParameterType(name string) string
}

type Product interface {
	GetDataModel() DataModel
}

type MethodCall interface {
	GetMethodName() string
	GetCommandKey() string
	GetRequestValues() map[string]string
	GetRequestValue(n string) string
}

type Device interface {
	GetProduct() Product
	GetNextMethodCall() MethodCall
	GetMethodCall(commandKey string) MethodCall
	PushMethodCall(t time.Time, methodName string, values map[string]any) (MethodCall, error)
	UpdateMethodCallRequestSend(commandKey string) error
	UpdateMethodCallResponse(commandKey string, values map[string]any, faultÇode int, faultString string) error
	UpdateMethodCallUnknow(commandKey string) error
}

type AcsHanlder interface {
	GetProduct(schema string, oui string, productClass string) Product
	GetDevice(schema string, oui string, productClass string, serialNumber string) Device

	HandleInform(ctx context.Context, inform *proto.Inform) error
	HandleFault(device Device, id string, v *proto.SoapFault) error

	HandleTransferComplete(ctx context.Context, device Device, req *proto.TransferComplete) error
	HandleAutonomousTransferComplete(ctx context.Context, device Device, req *proto.AutonomousTransferComplete) error
	HandleGetRPCMethodsResponse(ctx context.Context, device Device, id string, resp *proto.GetRPCMethodsResponse) error
	HandleGetParameterValuesResponse(ctx context.Context, device Device, id string, resp *proto.GetParameterValuesResponse) error
	HandleSetParameterValuesResponse(ctx context.Context, device Device, id string, resp *proto.SetParameterValuesResponse) error
	HandleGetParameterNamesResponse(ctx context.Context, device Device, id string, resp *proto.GetParameterNamesResponse) error
	HandleGetParameterAttributesResponse(ctx context.Context, device Device, id string, resp *proto.GetParameterAttributesResponse) error
	HandleSetParameterAttributesResponse(ctx context.Context, device Device, id string, resp *proto.SetParameterAttributesResponse) error
	HandleAddObjectResponse(ctx context.Context, device Device, id string, resp *proto.AddObjectResponse) error
	HandleDeleteObjectResponse(ctx context.Context, device Device, id string, resp *proto.DeleteObjectResponse) error
	HandleDownloadResponse(ctx context.Context, device Device, id string, resp *proto.DownloadResponse) error
	HandleUploadResponse(ctx context.Context, device Device, id string, resp *proto.UploadResponse) error
	HandleRebootResponse(ctx context.Context, device Device, id string, resp *proto.RebootResponse) error
	HandleFactoryResetResponse(ctx context.Context, device Device, id string, resp *proto.FactoryResetResponse) error

	HandleMesureValues(device Device, filename string, values map[string]any)
}
