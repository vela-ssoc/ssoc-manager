package restapi

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-common/tundata/mbreq"
	"github.com/vela-ssoc/ssoc-common/tundata/mbresp"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/brkclient"
	"github.com/vela-ssoc/ssoc-proto/muxconn"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
	"github.com/xgfone/ship/v5"
	"golang.org/x/time/rate"
)

type BrokerTunnel struct {
	hub muxserver.Huber
	cli brkclient.Client
}

func NewBrokerTunnel(hub muxserver.Huber, cli brkclient.Client) *BrokerTunnel {
	return &BrokerTunnel{
		hub: hub,
		cli: cli,
	}
}

func (bt *BrokerTunnel) RegisterRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/broker/tunnel/stat").GET(bt.stat)
	rgb.Route("/broker/tunnel/limit").POST(bt.limit)
	rgb.Route("/broker/tunnel/speedtest").GET(bt.speedtest)
	rgb.Route("/broker/tunnels").GET(bt.tunnels)
	rgb.Route("/broker/tunnel/kill").GET(bt.kill)

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

func (bt *BrokerTunnel) kill(c *ship.Context) error {
	req := new(request.BrokerTunnelKill)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	reqURL := muxproto.ManagerToBrokerURL(req.ID, "/api/v1/tunnel/kill")
	reqURL.RawQuery = c.Request().URL.RawQuery

	w, r := c.ResponseWriter(), c.Request()
	r.URL = reqURL

	base := bt.cli.Base() // http.
	base.ServeHTTP(w, r)

	return nil
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

	w, r := c.ResponseWriter(), c.Request()
	r.URL = reqURL

	base := bt.cli.Base() // http.
	base.ServeHTTP(w, r)

	return nil
}

func (bt *BrokerTunnel) tunnels(c *ship.Context) error {
	var ret []*mbresp.TunnelStat
	peers := bt.hub.Peers()
	for _, p := range peers {
		mux := p.MUX()
		name, module := mux.Library()
		cumulative, active := mux.NumStreams()
		tx, rx := mux.Traffic()
		bps := mux.Limit()
		ele := &mbresp.TunnelStat{
			Name:       name,
			Module:     module,
			Cumulative: cumulative,
			Active:     active,
			RX:         rx,
			TX:         tx,
			Limit:      float64(bps),
			Unlimit:    bps == rate.Inf,
		}
		for _, stm := range mux.Streams() {
			stats := stm.Stats()
			ele.Streams = append(ele.Streams, stats)
		}
		slices.SortFunc(ele.Streams, func(a, b *muxconn.StreamStats) int {
			return a.EstablishedAt.Compare(b.EstablishedAt)
		})

		ret = append(ret, ele)
	}

	return c.JSON(http.StatusOK, ret)
}
