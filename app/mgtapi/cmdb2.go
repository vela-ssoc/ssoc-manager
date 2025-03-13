package mgtapi

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Cmdb2(svc service.Cmdb2Service) route.Router {
	return &cmdb2REST{svc: svc}
}

type cmdb2REST struct {
	svc service.Cmdb2Service
}

func (rest *cmdb2REST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/cmdb2/minion/rsync").Data(route.Ignore()).GET(rest.Rsync)
}

func (rest *cmdb2REST) Rsync(c *ship.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	return rest.svc.Rsync(ctx)
}
