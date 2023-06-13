package mgtapi

import (
	"net"
	"net/http"
	"net/url"

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
		Data(route.Named("下载 agent 部署脚本")).GET(rest.Minion)
	anon.Route("/deploy/minion/download").
		Data(route.Named("下载 agent 二进制客户端")).GET(rest.MinionDownload)
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

func (rest *deployREST) Minion(c *ship.Context) error {
	var req param.DeployMinionDownload
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	r := c.Request()
	path := r.URL.Path
	downURL := &url.URL{
		Path:     path + "/download",
		RawQuery: r.URL.RawQuery,
	}
	return c.Redirect(http.StatusTemporaryRedirect, downURL.String())
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
