package mgtapi

import (
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

type logonREST struct {
	svc service.LogonService
}

func (lgn *logonREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	// TODO implement me
	panic("implement me")
}

func (lgn *logonREST) Page(c *ship.Context) error {
	return nil
}
