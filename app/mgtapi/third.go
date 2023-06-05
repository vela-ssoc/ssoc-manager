package mgtapi

import (
	"io"
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/xgfone/ship/v5"
)

func Third(svc service.ThirdService) route.Router {
	nameCol := dynsql.StringColumn("name", "文件名称").Build()
	descCol := dynsql.StringColumn("desc", "文件描述").Build()
	extCol := dynsql.StringColumn("extension", "文件后缀").Build()
	table := dynsql.Builder().
		Filters(nameCol, descCol, extCol).
		Build()

	return &thirdREST{
		svc:   svc,
		table: table,
	}
}

type thirdREST struct {
	svc   service.ThirdService
	table dynsql.Table
}

func (rest *thirdREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/third/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/thirds").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/third").
		Data(route.Ignore()).GET(rest.Download).
		Data(route.Named("新增三方文件")).POST(rest.Create).
		Data(route.Named("更新三方文件")).PUT(rest.Update).
		Data(route.Named("删除三方文件")).DELETE(rest.Delete)
}

func (rest *thirdREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *thirdREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *thirdREST) Create(c *ship.Context) error {
	var req param.ThirdCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	return rest.svc.Create(ctx, req.Name, req.Desc, file, cu.ID)
}

func (rest *thirdREST) Download(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	file, err := rest.svc.Download(ctx, req.ID)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	c.SetRespHeader(ship.HeaderContentDisposition, file.Disposition())
	c.SetRespHeader(ship.HeaderContentLength, file.ContentLength())

	return c.Stream(http.StatusOK, file.ContentType(), file)
}

func (rest *thirdREST) Delete(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}

func (rest *thirdREST) Update(c *ship.Context) error {
	var req param.ThirdUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	var r io.Reader
	if req.File != nil {
		file, err := req.File.Open()
		if err != nil {
			return err
		}
		//goland:noinspection GoUnhandledErrorResult
		defer file.Close()
		r = file
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return rest.svc.Update(ctx, req.ID, req.Desc, r, cu.ID)
}
