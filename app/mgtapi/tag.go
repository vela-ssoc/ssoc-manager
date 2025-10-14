package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewTag(svc service.TagService) *Tag {
	return &Tag{
		svc: svc,
	}
}

type Tag struct {
	svc service.TagService
}

func (rest *Tag) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/tag/indices").Data(route.Ignore()).GET(rest.Indices)
	bearer.Route("/tag/sidebar").Data(route.Ignore()).GET(rest.Sidebar)
	bearer.Route("/minion/tag").Data(route.Named("修改节点标签")).PATCH(rest.Update)
}

func (rest *Tag) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	idx := req.Indexer()
	res := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, res)
}

func (rest *Tag) Update(c *ship.Context) error {
	var req param.TagUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Update(ctx, req.ID, req.Tags)

	return err
}

func (rest *Tag) Sidebar(c *ship.Context) error {
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
