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

func Pprof(svc service.PprofService) route.Router {
	return &pprofREST{
		svc: svc,
	}
}

type pprofREST struct {
	svc service.PprofService
}

func (rest *pprofREST) Route(_, _, basic *ship.RouteGroupBuilder) {
	basic.Route("/flame/load").Data(route.Named("pprof-load")).GET(rest.Load)
	basic.Route("/flame/view").Data(route.Named("pprof-view")).GET(rest.View)
	basic.Route("/flame/view/*path").Data(route.Named("pprof-view")).GET(rest.View)
	basic.Route("/pprof/index").Data(route.Named("pprof-index")).GET(rest.Index)
	basic.Route("/pprof/cmdline").Data(route.Named("pprof-cmdline")).GET(rest.Cmdline)
	basic.Route("/pprof/profile").Data(route.Named("pprof-profile")).GET(rest.Profile)
	basic.Route("/pprof/symbol").Data(route.Named("pprof-symbol")).GET(rest.Symbol)
	basic.Route("/pprof/trace").Data(route.Named("pprof-trace")).GET(rest.Trace)
	basic.Route("/pprof/*path").Data(route.Named("pprof-path")).GET(rest.Path)
}

func (rest *pprofREST) Index(c *ship.Context) error {
	pprof.Index(c.Response(), c.Request())
	return nil
}

func (rest *pprofREST) Cmdline(c *ship.Context) error {
	pprof.Cmdline(c.Response(), c.Request())
	return nil
}

func (rest *pprofREST) Profile(c *ship.Context) error {
	pprof.Profile(c.Response(), c.Request())
	return nil
}

func (rest *pprofREST) Symbol(c *ship.Context) error {
	pprof.Symbol(c.Response(), c.Request())
	return nil
}

func (rest *pprofREST) Trace(c *ship.Context) error {
	pprof.Trace(c.Response(), c.Request())
	return nil
}

func (rest *pprofREST) Path(c *ship.Context) error {
	path := c.Param("path")
	pprof.Handler(path).ServeHTTP(c.Response(), c.Request())
	return nil
}

func (rest *pprofREST) Load(c *ship.Context) error {
	var req param.PprofLoad
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	second := req.Second
	if second <= 0 {
		second = 30
	}

	ctx := c.Request().Context()
	name, err := rest.svc.Load(ctx, req.Node, second)
	if err != nil {
		return err
	}
	res := &param.StrID{ID: name}

	return c.JSON(http.StatusOK, res)
}

func (rest *pprofREST) View(c *ship.Context) error {
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
	h, err := rest.svc.View(ctx, req.ID)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	w, r := c.Response(), c.Request()
	r.URL.Path = "/" + path
	h.ServeHTTP(w, r)

	return nil
}
