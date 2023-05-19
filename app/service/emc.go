package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
)

type EmcService interface{}

func Emc(hub linkhub.Huber, client netutil.HTTPClient) EmcService {
	return &emcService{}
}

type emcService struct {
	hub    linkhub.Huber
	client netutil.HTTPClient
}

func (emc *emcService) name() {
}

func (emc *emcService) Send(ctx context.Context, ids []string, title, content string) error {
	return nil
}
