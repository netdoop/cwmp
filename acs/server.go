package acs

import (
	"time"

	"go.uber.org/zap"
)

type AcsServer struct {
	logger              *zap.Logger
	dataRetentionPeriod time.Duration
	uploadBucket        string
	productStore        ProductStore
	deviceStore         DeviceStore
	pmHandler           PerformanceHanlder
}

func NewAcsServer(productStore ProductStore, deviceStore DeviceStore, pmHandler PerformanceHanlder, dataRetentionPeriod time.Duration) *AcsServer {
	s := AcsServer{
		logger:              zap.L().Named("acs"),
		productStore:        productStore,
		deviceStore:         deviceStore,
		pmHandler:           pmHandler,
		uploadBucket:        "acs-upload",
		dataRetentionPeriod: dataRetentionPeriod,
	}
	return &s
}

func (s *AcsServer) GetProduct(schema string, oui string, productClass string) Product {
	return s.productStore.GetProduct(schema, oui, productClass)
}

func (s *AcsServer) GetDevice(schema string, oui string, productClass string, serailNumber string) Device {
	return s.deviceStore.GetDevice(schema, oui, productClass, serailNumber)
}
