package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func NewUser(svc service.UserService) *User {
	return &User{svc: svc}
}

type User struct {
	svc service.UserService
}

func (rest *User) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/users").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/user/indices").Data(route.Ignore()).GET(rest.Indices)
	bearer.Route("/user/passwd").
		Data(route.Named("修改用户密码")).PATCH(rest.Passwd)
	bearer.Route("/user/sudo").
		Data(route.Named("修改用户资料")).PATCH(rest.Sudo)
	bearer.Route("/user").
		Data(route.Named("删除用户")).DELETE(rest.Delete).
		Data(route.Named("创建用户")).POST(rest.Create)
	bearer.Route("/user/ak").
		Data(route.Named("更新 AK")).PATCH(rest.AccessKey)
	bearer.Route("/user/totp").
		Data(route.Named("解绑 TOTP")).DELETE(rest.Totp)
}

// Page 分页查询
func (rest *User) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	page := req.Pager()
	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

// Indices 概要索引
func (rest *User) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	idx := req.Indexer()
	dats := rest.svc.Indices(ctx, idx)
	res := idx.Result(dats)

	return c.JSON(http.StatusOK, res)
}

// Sudo 超级管理员修改任意用户的信息
func (rest *User) Sudo(c *ship.Context) error {
	var req mrequest.UserSudo
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	logout, err := rest.svc.Sudo(ctx, &req)
	if err != nil || !logout {
		return err
	}

	if sess, ok := c.Session.(session.Session); ok {
		_ = sess.Destroy(req.ID)
	}

	return nil
}

func (rest *User) Create(c *ship.Context) error {
	var req mrequest.UserCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return rest.svc.Create(ctx, &req, cu.ID)
}

func (rest *User) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)
	if cu.ID == req.ID {
		return errcode.ErrDeleteSelf
	}

	if err := rest.svc.Delete(ctx, req.ID); err != nil {
		return err
	}
	if sess, ok := c.Session.(session.Session); ok {
		return sess.Destroy(req.ID)
	}

	return nil
}

func (rest *User) Passwd(c *ship.Context) error {
	var req mrequest.UserPasswd
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	err := rest.svc.Passwd(ctx, cu.ID, req.Original, req.Password)

	return err
}

func (rest *User) AccessKey(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.AccessKey(ctx, req.ID)

	return err
}

func (rest *User) Totp(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Totp(ctx, req.ID)

	return err
}
