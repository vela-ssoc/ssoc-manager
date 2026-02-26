package restapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-proto/muxconn"
	"github.com/xgfone/ship/v5"
)

func NewTunnel(acpt muxserver.MUXAccepter) *Tunnel {
	return &Tunnel{
		acpt: acpt,
		wsup: &websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
			CheckOrigin:      func(r *http.Request) bool { return true },
		},
	}
}

type Tunnel struct {
	acpt muxserver.MUXAccepter
	wsup *websocket.Upgrader
}

func (tnl *Tunnel) RegisterRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/tunnel").GET(tnl.open)

	return nil
}

// open broker 的接入点。
//
//goland:noinspection GoUnhandledErrorResult
func (tnl *Tunnel) open(c *ship.Context) error {
	proto := c.Query("protocol")
	if proto == "" {
		proto = c.Query("proto")
	}
	if proto != "yamux" {
		proto = "sumx"
	}

	attrs := []any{"protocol", proto, "remote_addr", c.RemoteAddr()}
	w, r := c.Response(), c.Request()
	ws, err := tnl.wsup.Upgrade(w, r, nil)
	if err != nil {
		attrs = append(attrs, "error", err)
		c.Warnf("升级为 websocket 协议出错", attrs...)
		return nil
	}

	var mux muxconn.Muxer
	ctx := context.Background()
	conn := ws.NetConn()
	if proto == "yamux" {
		mux, err = muxconn.NewYaMUX(ctx, conn, nil, true)
	} else {
		mux, err = muxconn.NewSMUX(ctx, conn, nil, true)
	}
	if err != nil {
		_ = ws.Close()
		attrs = append(attrs, "error", err)
		c.Warnf("升级为 muxer 协议出错", attrs...)
		return nil
	}

	c.Infof("通道接收到了新的连接", attrs...)
	tnl.acpt.AcceptMUX(mux)
	mux.Close()

	return nil
}
