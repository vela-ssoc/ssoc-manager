package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/xgfone/ship/v5"
)

func User(svc service.UserService) route.Router {
	return &userREST{svc: svc}
}

type userREST struct {
	svc service.UserService
}

func (usr *userREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/user/indices").Data(route.Ignore()).GET(usr.Indices)
	bearer.Route("/users").Data(route.Ignore()).GET(usr.Page)
}

// Page 分页查询
func (usr *userREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	page := req.Pager()
	count, dats := usr.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

// Indices 概要索引
func (usr *userREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	idx := req.Indexer()
	dats := usr.svc.Indices(ctx, idx)
	res := idx.Result(dats)

	return c.JSON(http.StatusOK, res)
}

// Sudo 超级管理员修改任意用户的信息
func (usr *userREST) Sudo(c *ship.Context) error {
	var req param.UserSudo
	if err := c.Bind(&req); err != nil {
		return err
	}

	if sess, ok := c.Session.(session.Session); ok {
		return sess.Destroy(req.ID)
	}

	return nil
}
