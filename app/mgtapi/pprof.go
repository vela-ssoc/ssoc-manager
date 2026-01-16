package mgtapi

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewPprof(svc *service.Pprof) *Pprof {
	return &Pprof{
		svc: svc,
	}
}

type Pprof struct {
	svc *service.Pprof
}

func (prf *Pprof) Route(_, _, basic *ship.RouteGroupBuilder) {
	basic.Route("/flame/load").Data(route.Named("pprof-load")).GET(prf.Load)
	basic.Route("/flame/view").Data(route.Named("pprof-view")).GET(prf.View)
	basic.Route("/flame/dump").Data(route.Named("pprof-dump")).GET(prf.dump)
	basic.Route("/flame/view/*path").Data(route.Named("pprof-view")).GET(prf.View)
	basic.Route("/pprof/index").Data(route.Named("pprof-index")).GET(prf.Index)
	basic.Route("/pprof/cmdline").Data(route.Named("pprof-cmdline")).GET(prf.Cmdline)
	basic.Route("/pprof/profile").Data(route.Named("pprof-profile")).GET(prf.Profile)
	basic.Route("/pprof/symbol").Data(route.Named("pprof-symbol")).GET(prf.Symbol)
	basic.Route("/pprof/trace").Data(route.Named("pprof-trace")).GET(prf.Trace)
	basic.Route("/pprof/*path").Data(route.Named("pprof-path")).GET(prf.Path)
}

func (prf *Pprof) Index(c *ship.Context) error {
	pprof.Index(c.Response(), c.Request())
	return nil
}

func (prf *Pprof) Cmdline(c *ship.Context) error {
	pprof.Cmdline(c.Response(), c.Request())
	return nil
}

func (prf *Pprof) Profile(c *ship.Context) error {
	pprof.Profile(c.Response(), c.Request())
	return nil
}

func (prf *Pprof) Symbol(c *ship.Context) error {
	pprof.Symbol(c.Response(), c.Request())
	return nil
}

func (prf *Pprof) Trace(c *ship.Context) error {
	pprof.Trace(c.Response(), c.Request())
	return nil
}

func (prf *Pprof) Path(c *ship.Context) error {
	path := c.Param("path")
	pprof.Handler(path).ServeHTTP(c.Response(), c.Request())
	return nil
}

func (prf *Pprof) Load(c *ship.Context) error {
	var req param.PprofLoad
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	name, err := prf.svc.Load(ctx, req.Node, req.Second)
	if err != nil {
		return err
	}
	res := &param.StrID{ID: name}

	return c.JSON(http.StatusOK, res)
}

func (prf *Pprof) View(c *ship.Context) error {
	var req param.StrID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	path := c.Param("path")
	rawURL := c.Request().URL
	rawPath := rawURL.Path
	if path == "" && !strings.HasSuffix(rawPath, "/") {
		return c.Redirect(http.StatusTemporaryRedirect, rawPath+"/"+"?"+rawURL.RawQuery)
	}

	ctx := c.Request().Context()
	h, err := prf.svc.View(ctx, req.ID)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	w, r := c.Response(), c.Request()
	r.URL.Path = "/" + path
	h.ServeHTTP(w, r)

	return nil
}

func (prf *Pprof) dump(c *ship.Context) error {
	req := new(param.PprofDump)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	name, err := prf.svc.Dump(ctx, req)
	if err != nil {
		return err
	}
	res := &param.StrID{ID: name}

	return c.JSON(http.StatusOK, res)
}
