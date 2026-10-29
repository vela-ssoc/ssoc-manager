package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type BrokerStream struct {
	svc *service.BrokerStream
}

func NewBrokerStream(svc *service.BrokerStream) *BrokerStream {
	return &BrokerStream{svc: svc}
}

func (bs *BrokerStream) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/broker/stream").DELETE(bs.close)       // 结束主连接
	g.ExposeV3.Group.Route("/broker/substreams").GET(bs.substreams) // 获取子流信息
	g.ExposeV3.Group.Route("/broker/substream").DELETE(bs.kill)     // 强制结束一个子流

	return nil
}

func (bs *BrokerStream) substreams(c *ship.Context) error {
	var param request.Slug
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	subs, _ := bs.svc.Substreams(param.Slug)

	var dats []response.BrokerSubstream
	for _, sub := range subs {
		ss := sub.Stats()
		rs := response.BrokerSubstream{
			ID:            ss.ID,
			RX:            ss.TX, // Manager 视角下，RX/TX 互换
			TX:            ss.RX, // Manager 视角下，RX/TX 互换
			EstablishedAt: ss.EstablishedAt,
		}
		dats = append(dats, rs)
	}
	ret := response.NewRecords(dats)

	return c.JSON(http.StatusOK, ret)
}

func (bs *BrokerStream) kill(c *ship.Context) error {
	var param request.BrokerSubstreamKill
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	var err error
	subs, _ := bs.svc.Substreams(param.Slug.Slug)
	for _, sub := range subs {
		ss := sub.Stats()
		if ss.ID == param.ID {
			err = sub.Close()
			break
		}
	}

	return err
}

func (bs *BrokerStream) close(c *ship.Context) error {
	var param request.Slug
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return bs.svc.Close(ctx, param.Slug)
}
