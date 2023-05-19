package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Tag(svc service.TagService) route.Router {
	return &tagREST{
		svc: svc,
	}
}

type tagREST struct {
	svc service.TagService
}

func (rest *tagREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/tag/indices").GET(rest.Indices)
}

func (rest *tagREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	idx := req.Indexer()
	ctx := c.Request().Context()
	res := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, res)
}
