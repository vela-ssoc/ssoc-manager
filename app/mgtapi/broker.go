package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/xgfone/ship/v5"
)

func Broker(svc service.BrokerService) route.Router {
	return &brokerREST{
		svc: svc,
	}
}

type brokerREST struct {
	svc service.BrokerService
}

func (rest *brokerREST) Route(anon, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/brokers").Data(route.Ignore()).GET(rest.Page)
	anon.Route("/broker/goos").Data(route.Ignore()).GET(rest.Goos)
	anon.Route("/broker/stats").Data(route.Ignore()).GET(rest.Stats)
	bearer.Route("/broker/indices").Data(route.Ignore()).GET(rest.Indices)
	bearer.Route("/broker").
		Data(route.Named("新增代理节点")).POST(rest.Create).
		Data(route.Named("修改代理节点")).PATCH(rest.Update).
		Data(route.Named("删除代理节点")).DELETE(rest.Delete)
}

func (rest *brokerREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *brokerREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	idx := req.Indexer()
	ctx := c.Request().Context()
	dats := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, dats)
}

func (rest *brokerREST) Create(c *ship.Context) error {
	var req mrequest.BrokerCreate
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.LAN) == 0 && len(req.VIP) == 0 {
		return errcode.ErrRequiredAddr
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *brokerREST) Update(c *ship.Context) error {
	var req mrequest.BrokerUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.LAN) == 0 && len(req.VIP) == 0 {
		return errcode.ErrRequiredAddr
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

func (rest *brokerREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}

func (rest *brokerREST) Goos(c *ship.Context) error {
	ctx := c.Request().Context()
	ret := rest.svc.Goos(ctx)
	return c.JSON(http.StatusOK, ret)
}

func (rest *brokerREST) Stats(c *ship.Context) error {
	ctx := c.Request().Context()
	res, err := rest.svc.Stats(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
