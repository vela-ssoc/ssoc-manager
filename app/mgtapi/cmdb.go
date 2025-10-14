package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewCmdb(svc *service.Cmdb) *Cmdb {
	return &Cmdb{svc: svc}
}

type Cmdb struct {
	svc *service.Cmdb
}

func (rest *Cmdb) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/cmdb").Data(route.Ignore()).GET(rest.Detail)
}

func (rest *Cmdb) Detail(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res := rest.svc.Detail(ctx, req.ID)

	return c.JSON(http.StatusOK, res)
}
