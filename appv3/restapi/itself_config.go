package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/xgfone/ship/v5"
)

type ItselfConfig struct {
	boot config.Config
	cfg  *model.ManagerConfig
	svc  *service.ItselfConfig
}

func NewItselfConfig(boot config.Config, cfg *model.ManagerConfig, svc *service.ItselfConfig) *ItselfConfig {
	return &ItselfConfig{boot: boot, cfg: cfg, svc: svc}
}

func (slf *ItselfConfig) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/itself/config/effect").GET(slf.effect) // 当前生效的配置参数
	g.ExposeV3.Group.Route("/itself/config").
		GET(slf.get).    // 获取最新配置参数（可能已保存但是还未生效）
		POST(slf.upsert) // 修改/新增配置

	return nil
}

func (slf *ItselfConfig) effect(c *ship.Context) error {
	ret := response.ItselfConfig{Boot: slf.boot, Config: slf.cfg}
	return c.JSON(http.StatusOK, ret)
}

func (slf *ItselfConfig) get(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, err := slf.svc.Get(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (slf *ItselfConfig) upsert(c *ship.Context) error {
	var param request.ItselfConfigUpdate
	if err := c.Bind(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return slf.svc.Upsert(ctx, param)
}
