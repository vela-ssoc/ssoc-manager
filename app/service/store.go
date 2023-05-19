package service

import (
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
)

type StoreService interface{}

type storeService struct {
	hub linkhub.Huber
}

func Store(hub linkhub.Huber) StoreService {
	return &storeService{
		hub: hub,
	}
}
