package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/application/expose/service"
	"github.com/xgfone/ship/v5"
)

type VictoriaMetricsConfig struct {
	svc *service.VictoriaMetricsConfig
}

func NewVictoriaMetricsConfig(svc *service.VictoriaMetricsConfig) *VictoriaMetricsConfig {
	return &VictoriaMetricsConfig{
		svc: svc,
	}
}

func (vmc *VictoriaMetricsConfig) BindRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/victoria-metrics-configs").
		Data(route.Named("查询 victoria-metrics 配置").Ignore()).POST(vmc.list)
	rgb.Route("/victoria-metrics-config").
		Data(route.Named("新增 victoria-metrics 配置")).POST(vmc.create).
		Data(route.Named("更新 victoria-metrics 配置")).PUT(vmc.update).
		Data(route.Named("删除 victoria-metrics 配置")).DELETE(vmc.delete)

	return nil
}

func (vmc *VictoriaMetricsConfig) list(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, err := vmc.svc.List(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (vmc *VictoriaMetricsConfig) create(c *ship.Context) error {
	req := new(request.VictoriaMetricsConfigCreate)
	ctx := c.Request().Context()
	if err := c.Bind(req); err != nil {
		return err
	}

	return vmc.svc.Create(ctx, req)
}

func (vmc *VictoriaMetricsConfig) update(c *ship.Context) error {
	req := new(request.VictoriaMetricsConfigUpdate)
	ctx := c.Request().Context()
	if err := c.Bind(req); err != nil {
		return err
	}

	return vmc.svc.Update(ctx, req)
}

func (vmc *VictoriaMetricsConfig) delete(c *ship.Context) error {
	req := new(request.Int64ID)
	ctx := c.Request().Context()
	if err := c.BindQuery(req); err != nil {
		return err
	}

	return vmc.svc.Delete(ctx, req.ID)
}
