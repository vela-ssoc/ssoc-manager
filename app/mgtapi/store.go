package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Store(svc service.StoreService) route.Router {
	return &storeREST{
		svc: svc,
	}
}

type storeREST struct {
	svc service.StoreService
}

func (rest *storeREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/store/indices").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/stores").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/store").
		Data(route.Ignore()).GET(rest.Detail).
		Data(route.Named("修改或新增模板")).POST(rest.Upsert).
		Data(route.Named("删除模板")).DELETE(rest.Delete)
}

func (rest *storeREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

// Detail 查询单个配置数据
func (rest *storeREST) Detail(c *ship.Context) error {
	id := c.Query("id")
	ctx := c.Request().Context()
	res, err := rest.svc.FindID(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (rest *storeREST) Upsert(c *ship.Context) error {
	var req param.StoreUpsert
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Upsert(ctx, &req)
}

func (rest *storeREST) Delete(c *ship.Context) error {
	id := c.Query("id")
	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, id)
}
