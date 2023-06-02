package mgtapi

import (
	"net/http"
	"strings"

	"github.com/vela-ssoc/vela-common-mb/davfs"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

func DavFS(base string) route.Router {
	base = strings.TrimRight(base, "/")
	still := "/dav"
	prefix := base + still

	fs := davfs.FS("/", prefix)

	return &davREST{
		still: still,
		dav:   fs,
	}
}

type davREST struct {
	still string
	dav   http.Handler
}

func (rest *davREST) Route(_, _, basic *ship.RouteGroupBuilder) {
	allows := []string{
		http.MethodOptions, http.MethodGet, http.MethodHead, http.MethodPost, "LOCK", "UNLOCK", "PROPFIND",
	}
	forbids := []string{http.MethodPut, http.MethodDelete, "PROPPATCH", "MKCOL", "COPY", "MOVE"}

	basic.Route(rest.still).
		Method(rest.DAV, allows...).
		Method(rest.Forbidden, forbids...)
	basic.Route(rest.still+"/*path").
		Method(rest.DAV, allows...).
		Method(rest.Forbidden, forbids...)
}

func (rest *davREST) DAV(c *ship.Context) error {
	// path := "/" + c.Param("path")
	w, r := c.Response(), c.Request()
	// r.URL.Path = path
	rest.dav.ServeHTTP(w, r)
	return nil
}

func (rest *davREST) Forbidden(*ship.Context) error {
	return errcode.ErrForbidden
}
