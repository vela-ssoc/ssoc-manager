package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type ReverseProxy struct {
	svc *service.ReverseProxy
}

func NewReverseProxy(svc *service.ReverseProxy) *ReverseProxy {
	return &ReverseProxy{svc: svc}
}

func (rp *ReverseProxy) RegisterHTTP(g shipx.RouteGroup) error {
	g.APIV3.Route("/reverse-proxies").GET(rp.list)    // 列表显示
	g.APIV3.Route("/reverse-proxy").GET(rp.create)    // 单个信息
	g.APIV3.Route("/reverse-proxy").POST(rp.create)   // 新增
	g.APIV3.Route("/reverse-proxy").PUT(rp.update)    // 修改
	g.APIV3.Route("/reverse-proxy").DELETE(rp.delete) // 删除

	g.APIV3NoAuth.Route("/reverse/:location").Any(rp.handle)
	g.APIV3NoAuth.Route("/reverse/:location/*path").Any(rp.handle)

	return nil
}

func (rp *ReverseProxy) list(c *ship.Context) error {
	ctx := c.Request().Context()
	dats, err := rp.svc.List(ctx)
	if err != nil {
		return err
	}
	ret := response.NewRecords(dats)

	return c.JSON(http.StatusOK, ret)
}

func (rp *ReverseProxy) get(c *ship.Context) error {
	var param request.ReverseProxyLocation
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := rp.svc.Get(ctx, param.Location)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (rp *ReverseProxy) create(c *ship.Context) error {
	var param request.ReverseProxyUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rp.svc.Create(ctx, param)
}

func (rp *ReverseProxy) update(c *ship.Context) error {
	var param request.ReverseProxyUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rp.svc.Update(ctx, param)
}

func (rp *ReverseProxy) delete(c *ship.Context) error {
	var param request.ReverseProxyLocation
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rp.svc.Delete(ctx, param.Location)
}

func (rp *ReverseProxy) handle(c *ship.Context) error {
	w, r := shipx.Strip(c, "path")
	ctx := r.Context()
	loc := c.Param("location")
	r.Header.Del(ship.HeaderAuthorization)

	if err := rp.svc.Handle(ctx, loc, w, r); err != nil {
		return ship.ErrNotFound
	}

	return nil
}
