package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type AgentConnectHistory struct {
	svc *service.AgentConnectHistory
}

func NewAgentConnectHistory(svc *service.AgentConnectHistory) *AgentConnectHistory {
	return &AgentConnectHistory{svc: svc}
}

func (ach *AgentConnectHistory) RegisterHTTP(g shipx.RouteGroup) error {
	g.ExposeV3.Group.Route("/agent/connect-histories").GET(ach.page)

	return nil
}

func (ach *AgentConnectHistory) page(c *ship.Context) error {
	var param request.AgentConnectHistoryPage
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := ach.svc.Page(ctx, param.ID, param.PageSize)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
