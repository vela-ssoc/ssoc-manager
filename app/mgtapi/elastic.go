package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Elastic(svc service.ElasticService, headerKey string, queryKey string) route.Router {
	return &elasticREST{
		svc:       svc,
		headerKey: headerKey,
		queryKey:  queryKey,
	}
}

type elasticREST struct {
	svc       service.ElasticService
	headerKey string
	queryKey  string
}

func (ela *elasticREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/elastics").Data(route.Ignore()).GET(ela.Page)
	bearer.Route("/ribana/*path").Data(route.Named("Ribana")).Any(ela.Forward)
	bearer.Route("/ribana").Data(route.Named("Ribana")).Any(ela.Forward)
	bearer.Route("/elastic").
		Data(route.Named("新增 es 服务器")).POST(ela.Create).
		Data(route.Named("修改 es 服务器")).PUT(ela.Update).
		Data(route.Named("删除 es 服务器")).DELETE(ela.Delete)
}

func (ela *elasticREST) Forward(c *ship.Context) error {
	path := "/" + c.Param("path")
	w, r := c.Response(), c.Request()
	ctx := r.Context()
	r.URL.Path = path

	ela.desensitization(r)

	return ela.svc.Forward(ctx, w, r)
}

func (ela *elasticREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := ela.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (ela *elasticREST) Create(c *ship.Context) error {
	var req param.ElasticCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return ela.svc.Create(ctx, &req)
}

func (ela *elasticREST) Update(c *ship.Context) error {
	var req param.ElasticUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := ela.svc.Update(ctx, &req)

	return err
}

func (ela *elasticREST) Delete(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return ela.svc.Delete(ctx, req.ID)
}

// desensitization 对代理转发的请求脱敏，前端用户请求携带的认证信息不应该带到后面的节点请求中。
func (ela *elasticREST) desensitization(r *http.Request) {
	r.Header.Del(ela.headerKey)
	query := r.URL.Query()
	query.Del(ela.queryKey)
	r.URL.RawQuery = query.Encode()
}
