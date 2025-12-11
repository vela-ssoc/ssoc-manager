package restapi

import (
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/application/expose/service"
	"github.com/xgfone/ship/v5"
)

func NewSubstanceExtension(svc *service.SubstanceExtension) *SubstanceExtension {
	return &SubstanceExtension{
		svc: svc,
	}
}

type SubstanceExtension struct {
	svc *service.SubstanceExtension
}

func (se *SubstanceExtension) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/substance/extension").
		POST(se.create).
		PUT(se.update)
}

func (se *SubstanceExtension) create(c *ship.Context) error {
	req := new(request.SubstanceExtensionCreate)
	if err := c.Bind(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return se.svc.Create(ctx, req, cu)
}

func (se *SubstanceExtension) update(c *ship.Context) error {
	req := new(request.SubstanceExtensionUpdate)
	if err := c.Bind(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return se.svc.Update(ctx, req, cu)
}
