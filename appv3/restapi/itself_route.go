package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type ItselfRoute struct {
	brk *ship.Ship
	exp *ship.Ship
}

func NewItselfRoute() *ItselfRoute {
	return new(ItselfRoute)
}

func (rt *ItselfRoute) RegisterHTTP(g shipx.RouteGroup3) error {
	rt.exp = g.ExposeV3.Group.Ship()
	rt.brk = g.BrokerV3.Group.Ship()

	g.ExposeV3.Group.Route("/itself/route/brokers").GET(rt.brokers) // 中继节点侧接口
	g.ExposeV3.Group.Route("/itself/route/exposes").GET(rt.exposes) // 用户侧接口

	return nil
}

func (rt *ItselfRoute) brokers(c *ship.Context) error {
	ret := rt.foreach(rt.brk)
	return c.JSON(http.StatusOK, ret)
}

func (rt *ItselfRoute) exposes(c *ship.Context) error {
	ret := rt.foreach(rt.exp)
	return c.JSON(http.StatusOK, ret)
}

func (*ItselfRoute) foreach(sh *ship.Ship) response.Records[response.RouteInfo] {
	dats := make([]response.RouteInfo, 0, 50)
	if sh != nil && sh.Router != nil {
		sh.Router.Range(func(_, path, method string, _ any) {
			item := response.RouteInfo{Method: method, Path: path}
			dats = append(dats, item)
		})
	}

	return response.NewRecords(dats)
}
