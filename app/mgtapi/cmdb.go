package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Cmdb(svc service.CmdbService) route.Router {
	return &cmdbREST{svc: svc}
}

type cmdbREST struct {
	svc service.CmdbService
}

func (rest *cmdbREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/cmdb").GET(rest.Detail)
}

func (rest *cmdbREST) Detail(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res := rest.svc.Detail(ctx, req.ID)

	return c.JSON(http.StatusOK, res)
}
