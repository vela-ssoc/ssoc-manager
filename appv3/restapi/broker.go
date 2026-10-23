package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type Broker struct {
	svc *service.Broker
}

func NewBroker(svc *service.Broker) *Broker {
	return &Broker{svc: svc}
}

func (bk *Broker) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/brokers").GET(bk.list)     // 中继节点列表
	g.ExposeV3.Group.Route("/broker").GET(bk.get)       // 获取中继节点详情
	g.ExposeV3.Group.Route("/broker").POST(bk.create)   // 创建中继节点
	g.ExposeV3.Group.Route("/broker").PUT(bk.update)    // 修改中继节点
	g.ExposeV3.Group.Route("/broker").DELETE(bk.delete) // 删除中继节点
	g.BrokerV3.Group.Route("/broker/ping").GET(bk.ping) // 中继节点发送心跳包

	return nil
}

func (bk *Broker) get(c *ship.Context) error {
	var param request.Slug
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := bk.svc.Get(ctx, param.Slug)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (bk *Broker) list(c *ship.Context) error {
	ctx := c.Request().Context()
	dats, err := bk.svc.List(ctx)
	if err != nil {
		return err
	}
	ret := response.NewRecords(dats)

	return c.JSON(http.StatusOK, ret)
}

func (bk *Broker) create(c *ship.Context) error {
	var param request.BrokerUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return bk.svc.Create(ctx, param)
}

func (bk *Broker) update(c *ship.Context) error {
	var param request.BrokerUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return bk.svc.Update(ctx, param)
}

func (bk *Broker) delete(c *ship.Context) error {
	var param request.Slug
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return bk.svc.Delete(ctx, param.Slug)
}

func (bk *Broker) ping(c *ship.Context) error {
	ctx := c.Request().Context()
	peer := muxproto.FromContext(ctx)

	return bk.svc.Ping(ctx, peer)
}
