package mgtapi

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewCmdb2(svc *service.Cmdb2) *Cmdb2 {
	return &Cmdb2{svc: svc}
}

type Cmdb2 struct {
	svc *service.Cmdb2
}

func (rest *Cmdb2) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/cmdb2/minion/rsync").Data(route.Ignore()).GET(rest.Rsync)
}

func (rest *Cmdb2) Rsync(c *ship.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	return rest.svc.Rsync(ctx)
}
