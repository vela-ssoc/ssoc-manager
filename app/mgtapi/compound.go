package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Compound(svc service.CompoundService) route.Router {
	return &compoundREST{
		svc: svc,
	}
}

type compoundREST struct {
	svc service.CompoundService
}

func (rest *compoundREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/compound/indices").Data(route.Ignore()).GET(rest.Indices)
}

func (rest *compoundREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	idx := req.Indexer()
	ctx := c.Request().Context()
	res := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, res)
}
