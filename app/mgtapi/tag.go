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
	bearer.Route("/tag/indices").Data(route.Ignore()).GET(rest.Indices)
	bearer.Route("/tag/sidebar").Data(route.Ignore()).GET(rest.Sidebar)
	bearer.Route("/minion/tag").Data(route.Named("修改节点标签")).PATCH(rest.Update)
}

func (rest *tagREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	idx := req.Indexer()
	res := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, res)
}

func (rest *tagREST) Update(c *ship.Context) error {
	var req param.TagUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Update(ctx, req.ID, req.Tags)

	return err
}

func (rest *tagREST) Sidebar(c *ship.Context) error {
	req := new(param.TagSidebar)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := rest.svc.Sidebar(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
