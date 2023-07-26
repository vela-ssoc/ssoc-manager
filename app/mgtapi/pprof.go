package mgtapi

import (
	"net/http/pprof"

	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/xgfone/ship/v5"
)

func Pprof() route.Router {
	return &pprofREST{}
}

type pprofREST struct{}

func (rest *pprofREST) Route(_, _, basic *ship.RouteGroupBuilder) {
	basic.Route("/debug/pprof").Data(route.Named("pprof-index")).GET(rest.Index)
	basic.Route("/debug/cmdline").Data(route.Named("pprof-cmdline")).GET(rest.Cmdline)
	basic.Route("/debug/profile").Data(route.Named("pprof-profile")).GET(rest.Profile)
	basic.Route("/debug/symbol").Data(route.Named("pprof-symbol")).GET(rest.Symbol)
	basic.Route("/debug/trace").Data(route.Named("pprof-trace")).GET(rest.Trace)
	basic.Route("/debug/*path").Data(route.Named("pprof-path")).GET(rest.Path)
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
