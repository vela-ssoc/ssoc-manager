package mgtapi

import (
	"bytes"
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewLuaTemplate(svc *service.LuaTemplate) *LuaTemplate {
	return &LuaTemplate{
		svc: svc,
	}
}

type LuaTemplate struct {
	svc *service.LuaTemplate
}

func (lt *LuaTemplate) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/lua-template/preparse").Data(route.Ignore()).POST(lt.preparse)
	bearer.Route("/lua-template/prerender").Data(route.Ignore()).POST(lt.prerender)
}

func (lt *LuaTemplate) preparse(c *ship.Context) error {
	req := new(param.LuaTemplatePreparse)
	if err := c.Bind(req); err != nil {
		return err
	}

	dat, err := lt.svc.Preparse(req.Content)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dat)
}

func (lt *LuaTemplate) prerender(c *ship.Context) error {
	req := new(param.LuaTemplatePrerender)
	if err := c.Bind(req); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	source, data := req.Content, req.Data
	if err := lt.svc.Prerender(buf, source, data); err != nil {
		return err
	}
	dat := &param.Data{Data: buf.String()}

	return c.JSON(http.StatusOK, dat)
}
