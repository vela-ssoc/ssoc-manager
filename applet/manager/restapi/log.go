package restapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/applet/manager/request"
	"github.com/vela-ssoc/vela-manager/applet/manager/response"
	"github.com/vela-ssoc/vela-manager/applet/manager/service"
	"github.com/xgfone/ship/v5"
)

func NewLog(svc *service.Log) *Log {
	return &Log{
		svc: svc,
	}
}

type Log struct {
	svc *service.Log
}

func (lg *Log) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/log/levels").GET(lg.levels)
	r.Route("/log/level").POST(lg.change)
	return nil
}

func (lg *Log) levels(c *ship.Context) error {
	l1, l2 := lg.svc.Levels()
	ret := &response.LogLevel{Log: l1, ORM: l2}

	return c.JSON(http.StatusOK, ret)
}

func (lg *Log) change(c *ship.Context) error {
	req := new(request.LogChange)
	if err := c.Bind(req); err != nil {
		return err
	}
	lg.svc.Change(req)

	return nil
}
