package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/xgfone/ship/v5"
)

func NewExtensionMarket(svc *service.ExtensionMarket) *ExtensionMarket {
	return &ExtensionMarket{
		svc: svc,
	}
}

type ExtensionMarket struct {
	svc *service.ExtensionMarket
}

func (mkt *ExtensionMarket) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/extension-markets").Data(route.Ignore()).GET(mkt.page)
	bearer.Route("/extension-market/records").Data(route.Ignore()).GET(mkt.records)
	bearer.Route("/extension-market").Data(route.Ignore()).POST(mkt.create)
	bearer.Route("/extension-market").Data(route.Ignore()).PUT(mkt.update)
	bearer.Route("/extension-market").Data(route.Ignore()).DELETE(mkt.delete)
}

func (mkt *ExtensionMarket) page(c *ship.Context) error {
	req := new(param.ExtensionMarketPage)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	page := req.Pager()
	cnt, dats := mkt.svc.Page(ctx, page, req.Category)
	dat := page.Result(cnt, dats)

	return c.JSON(http.StatusOK, dat)
}

func (mkt *ExtensionMarket) create(c *ship.Context) error {
	req := new(param.ExtensionMarketCreate)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	ret, err := mkt.svc.Create(ctx, req, cu)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (mkt *ExtensionMarket) update(c *ship.Context) error {
	req := new(param.ExtensionMarketUpdate)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return mkt.svc.Update(ctx, req, cu)
}

func (mkt *ExtensionMarket) delete(c *ship.Context) error {
	req := new(param.IntID)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return mkt.svc.Delete(ctx, req.ID)
}

func (mkt *ExtensionMarket) records(c *ship.Context) error {
	req := new(param.IntID)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	dat, err := mkt.svc.Records(ctx, req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dat)
}
