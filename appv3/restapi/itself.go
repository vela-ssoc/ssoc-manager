package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/bininfo"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type Itself struct {
}

func NewItself() *Itself {
	return &Itself{}
}

func (slf *Itself) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/itself/buildinfo").GET(slf.buildinfo) // 二进制编译信息
	return nil
}

func (slf *Itself) buildinfo(c *ship.Context) error {
	ret := bininfo.Self()
	return c.JSON(http.StatusOK, ret)
}
