package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func NewCert(svc service.CertService) *Cert {
	return &Cert{
		svc: svc,
	}
}

type Cert struct {
	svc service.CertService
}

func (rest *Cert) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/certs").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/cert/indices").Data(route.Ignore()).GET(rest.Indices)
	bearer.Route("/cert").
		Data(route.Named("添加证书")).POST(rest.Create).
		Data(route.Named("修改证书")).PUT(rest.Update).
		Data(route.Named("删除证书")).DELETE(rest.Delete)
}

func (rest *Cert) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	pager := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, pager)
	res := pager.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *Cert) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	idx := req.Indexer()
	ctx := c.Request().Context()
	res := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, res)
}

func (rest *Cert) Create(c *ship.Context) error {
	var req mrequest.CertCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *Cert) Update(c *ship.Context) error {
	var req mrequest.CertUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

func (rest *Cert) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
