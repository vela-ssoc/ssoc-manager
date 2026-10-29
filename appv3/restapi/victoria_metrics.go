package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type VictoriaMetrics struct {
	svc *service.VictoriaMetrics
}

func NewVictoriaMetrics(svc *service.VictoriaMetrics) *VictoriaMetrics {
	return &VictoriaMetrics{svc: svc}
}

func (vmc *VictoriaMetrics) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/victoria-metrics-config").
		GET(vmc.get).    // 获取 VictoriaMetrics 配置
		POST(vmc.upsert) // 修改/新增 VictoriaMetrics 配置
	g.ExposeV3.Group.Route("/vmp/*path").Any(vmc.reverse)

	return nil
}

func (vmc *VictoriaMetrics) get(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, _ := vmc.svc.Get(ctx)

	return c.JSON(http.StatusOK, ret)
}

func (vmc *VictoriaMetrics) upsert(c *ship.Context) error {
	var param request.VictoriaMetricsConfigUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return vmc.svc.Upsert(ctx, param)
}

func (vmc *VictoriaMetrics) reverse(c *ship.Context) error {
	w, r := shipx.Strip(c, "path")
	ctx := r.Context()

	return vmc.svc.Reverse(ctx, w, r)
}
