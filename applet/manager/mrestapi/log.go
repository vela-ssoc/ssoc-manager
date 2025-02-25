package mrestapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/applet/manager/mrequest"
	"github.com/vela-ssoc/vela-manager/applet/manager/mresponse"
	"github.com/vela-ssoc/vela-manager/applet/manager/mservice"
	"github.com/xgfone/ship/v5"
)

func NewLog(svc *mservice.Log) *Log {
	return &Log{
		svc: svc,
	}
}

type Log struct {
	svc *mservice.Log
}

func (lg *Log) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/log/levels").GET(lg.levels)
	r.Route("/log/level").POST(lg.change)
	return nil
}

func (lg *Log) levels(c *ship.Context) error {
	l1, l2 := lg.svc.Levels()
	ret := &mresponse.LogLevel{Log: l1, ORM: l2}

	return c.JSON(http.StatusOK, ret)
}

func (lg *Log) change(c *ship.Context) error {
	req := new(mrequest.LogChange)
	if err := c.Bind(req); err != nil {
		return err
	}
	lg.svc.Change(req)

	return nil
}
