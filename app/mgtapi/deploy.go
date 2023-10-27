package mgtapi

import (
	"net"
	"net/http"
	"net/url"

	"github.com/vela-ssoc/vela-manager/app/internal/modview"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Deploy(svc service.DeployService) route.Router {
	return &deployREST{
		svc: svc,
	}
}

type deployREST struct {
	svc service.DeployService
}

func (rest *deployREST) Route(anon, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/deploy/lan").Data(route.Ignore()).GET(rest.LAN)
	anon.Route("/deploy/minion").
		Data(route.Ignore()).GET(rest.Script)
	anon.Route("/deploy/minion/download").
		Data(route.Ignore()).GET(rest.MinionDownload)
}

func (rest *deployREST) LAN(c *ship.Context) error {
	res := &param.DeployLAN{Scheme: "http"}
	r := c.Request()
	if r.TLS != nil {
		res.Scheme = "https"
	}

	ctx := r.Context()
	if addr := rest.svc.LAN(ctx); addr != "" {
		res.Addr = addr
		return c.JSON(http.StatusOK, res)
	}

	val := ctx.Value(http.LocalAddrContextKey)
	if ip, ok := val.(net.Addr); ok {
		res.Addr = ip.String()
	} else {
		res.Addr = r.Host
	}

	return c.JSON(http.StatusOK, res)
}

func (rest *deployREST) Script(c *ship.Context) error {
	var req param.DeployMinionDownload
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	r := c.Request()
	reqURL := r.URL

	scheme := "http"
	if c.IsTLS() {
		scheme = "https"
	}
	// 如果 TLS 证书挂在了 WAF 上
	proto := c.GetReqHeader(ship.HeaderXForwardedProto)
	if proto == "http" || proto == "https" {
		scheme = proto
	}

	path := reqURL.Path + "/download"
	downURL := &url.URL{
		Scheme:   scheme,
		Host:     r.Host,
		Path:     path,
		RawQuery: reqURL.RawQuery,
	}
	ctx := c.Request().Context()

	data := &modview.Deploy{DownloadURL: downURL}
	read, err := rest.svc.Script(ctx, req.Goos, data)
	if err == nil {
		return c.Stream(http.StatusOK, ship.MIMETextPlainCharsetUTF8, read)
	}

	redirectURL := &url.URL{
		Path:     path,
		RawQuery: r.URL.RawQuery,
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
}

func (rest *deployREST) MinionDownload(c *ship.Context) error {
	var req param.DeployMinionDownload
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	file, err := rest.svc.OpenMinion(ctx, &req)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	// 此时的 Content-Length = 原始文件 + 隐藏文件
	c.Header().Set(ship.HeaderContentLength, file.ContentLength())
	c.Header().Set(ship.HeaderContentDisposition, file.Disposition())

	return c.Stream(http.StatusOK, file.ContentType(), file)
}
