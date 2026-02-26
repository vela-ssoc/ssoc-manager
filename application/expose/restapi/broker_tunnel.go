package restapi

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/vela-ssoc/ssoc-common/tundata/mbreq"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/brkclient"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
	"github.com/xgfone/ship/v5"
	"golang.org/x/time/rate"
)

type BrokerTunnel struct {
	cli brkclient.Client
}

func NewBrokerTunnel(cli brkclient.Client) *BrokerTunnel {
	return &BrokerTunnel{
		cli: cli,
	}
}

func (bt *BrokerTunnel) RegisterRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/broker/tunnel/stat").GET(bt.stat)
	rgb.Route("/broker/tunnel/limit").POST(bt.limit)
	rgb.Route("/broker/tunnel/speedtest").GET(bt.speedtest)

	return nil
}

func (bt *BrokerTunnel) stat(c *ship.Context) error {
	req := new(request.HexID)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := bt.cli.TunnelStat(ctx, req.MustID())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (bt *BrokerTunnel) limit(c *ship.Context) error {
	req := new(request.BrokerTunnelLimit)
	if err := c.Bind(req); err != nil {
		return err
	}

	limit := float64(req.Limit)
	ctx := c.Request().Context()
	dat := &mbreq.TunnelLimit{
		Unlimit: rate.Limit(limit) == rate.Inf,
		Limit:   limit,
	}

	return bt.cli.TunnelLimit(ctx, req.MustID(), dat)
}

func (bt *BrokerTunnel) speedtest(c *ship.Context) error {
	req := new(request.BrokerTunnelSpeedtest)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	num := strconv.FormatInt(int64(req.Size), 10)
	quires := make(url.Values, 2)
	quires.Set("size", num)

	reqURL := muxproto.ManagerToBrokerURL(req.ID, "/api/v1/speedtest")
	reqURL.RawQuery = quires.Encode()

	r := c.Request()
	r.URL = reqURL

	base := bt.cli.Base()
	base.ServeHTTP(c, r)

	return nil
}
