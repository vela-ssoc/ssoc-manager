package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

type oplogREST struct {
	svc   service.OplogService
	table dynsql.Table
}

func Oplog(svc service.OplogService) route.Router {
	methods := []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace,
	}
	methodEnum := dynsql.StringEnum().Sames(methods)

	idCol := dynsql.StringColumn("id", "ID").Build()
	methodCol := dynsql.StringColumn("method", "请求方法").Enums(methodEnum).Build()
	nicknameCol := dynsql.StringColumn("nickname", "用户名").Build()

	table := dynsql.Builder().
		Filters(idCol, methodCol, nicknameCol).
		Build()

	return &oplogREST{
		svc:   svc,
		table: table,
	}
}

func (op *oplogREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/oplog/cond").Data(route.Ignore()).GET(op.Cond)
	bearer.Route("/oplogs").Data(route.Ignore()).GET(op.Page)
	bearer.Route("/oplog").Data(route.Named("删除操作日志")).DELETE(op.Delete)
}

func (op *oplogREST) Cond(c *ship.Context) error {
	res := op.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (op *oplogREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := op.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := op.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (op *oplogREST) Delete(c *ship.Context) error {
	var req dynsql.Input
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.Filters) == 0 {
		return errcode.ErrRequiredFilter
	}
	scope, err := op.table.Inter(req)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	ctx := c.Request().Context()

	return op.svc.Delete(ctx, scope)
}
