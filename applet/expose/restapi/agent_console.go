package restapi

import (
	"github.com/vela-ssoc/ssoc-manager/applet/expose/request"
	"github.com/vela-ssoc/ssoc-manager/applet/expose/service"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/xgfone/ship/v5"
)

func NewAgentConsole(hub linkhub.Huber, svc *service.Agent) *AgentConsole {
	return &AgentConsole{
		hub: hub,
		svc: svc,
	}
}

type AgentConsole struct {
	hub linkhub.Huber
	svc *service.Agent
}

func (ac *AgentConsole) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/agent/console/read").GET(ac.read)
}

func (ac *AgentConsole) read(c *ship.Context) error {
	req := new(request.AgentConsoleRead)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	// TODO 查询 agent 所在的 broker 节点
	w, r := c.Response(), c.Request()
	ctx := r.Context()
	mon, err := ac.svc.Get(ctx, req.ID)
	if err != nil || mon.BrokerID == 0 {
		return err
	}

	ac.hub.Forward(mon.BrokerID, w, r)

	return nil
}
