package restapi

import (
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/vela-ssoc/ssoc-common/linkhub"
	"github.com/xgfone/ship/v5"
)

type Reverse struct {
	pxy *httputil.ReverseProxy
}

func NewReverse(tran http.RoundTripper) *Reverse {
	pxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
		},
		Director:      nil,
		Transport:     tran,
		FlushInterval: 0,
		ErrorLog:      nil,
		BufferPool:    nil,
		ModifyResponse: func(w *http.Response) error {
			if w.StatusCode == http.StatusUnauthorized {
				w.StatusCode = http.StatusBadRequest
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
		},
	}

	return &Reverse{
		pxy: pxy,
	}
}

func (rvs *Reverse) BindRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/reverse/*path").Any(rvs.serve)
	rgb.Route("/reverse/agent/console/read").GET(rvs.serve)

	return nil
}

func (rvs *Reverse) serve(c *ship.Context) error {
	w, r := c.Response(), c.Request()
	path := "/" + c.Param("path")
	if path != "/" && strings.HasSuffix(r.URL.Path, "/") {
		path += "/"
	}
	reqURL := linkhub.NewServerToBrokerURL("", path)
	reqURL.RawQuery = r.URL.RawQuery

	rvs.pxy.ServeHTTP(w, r)

	return nil
}
