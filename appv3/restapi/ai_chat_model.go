package restapi

import (
	"context"
	"net/http"
	"time"

	"github.com/meguminnnnnnnnn/go-openai"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type AIChatModel struct {
	svc     *service.AIChatModel
	httpcli *http.Client
}

func NewAIChatModel(svc *service.AIChatModel, httpcli *http.Client) *AIChatModel {
	return &AIChatModel{svc: svc, httpcli: httpcli}
}

func (acm *AIChatModel) RegisterHTTP(g shipx.RouteGroup) error {
	g.ExposeV3.Group.Route("/ai-chat-model").
		GET(acm.get).    // 获取聊天模型配置
		POST(acm.upsert) // 修改/新增聊天模型配置
	g.ExposeV3.Group.Route("/ai-chat-model/models").
		GET(acm.models).     // 获取可用的聊天模型列表
		POST(acm.listModels) // 获取上游服务的模型列表

	return nil
}

func (acm *AIChatModel) get(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, _ := acm.svc.Get(ctx)

	return c.JSON(http.StatusOK, ret)
}

func (acm *AIChatModel) upsert(c *ship.Context) error {
	var param request.AIChatModelUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return acm.svc.Upsert(ctx, param)
}

func (acm *AIChatModel) models(c *ship.Context) error {
	ctx := c.Request().Context()
	dat, _ := acm.svc.Get(ctx)
	var ret response.Records[string]
	if dat != nil && dat.Enabled {
		ret.Records = dat.Models
	}

	return c.JSON(http.StatusOK, ret)
}

func (acm *AIChatModel) listModels(c *ship.Context) error {
	var param request.AIChatModelModels
	if err := c.Bind(&param); err != nil {
		return err
	}

	parent := c.Request().Context()
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()

	cfg := openai.DefaultConfig(param.APIKey)
	cfg.BaseURL = param.BaseURL
	cfg.HTTPClient = acm.httpcli

	cli := openai.NewClientWithConfig(cfg)
	ret, err := cli.ListModels(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
