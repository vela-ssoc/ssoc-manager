package restapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/vela-tunnel/muxconn"
	"github.com/xgfone/ship/v5"
)

type Tunnel struct {
	han muxproto.MUXServer
	wsu *websocket.Upgrader
}

func NewTunnel(han muxproto.MUXServer) *Tunnel {
	return &Tunnel{
		han: han,
		wsu: &websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
			Subprotocols:     []string{"smux", "yamux", "vela"},
			CheckOrigin:      func(*http.Request) bool { return true },
		},
	}
}

func (tnl *Tunnel) RegisterHTTP(g shipx.RouteGroup) error {
	g.APIV3NoAuth.Route("/tunnel").GET(tnl.handle) // 中继节点连接端点

	return nil
}

func (tnl *Tunnel) handle(c *ship.Context) error {
	w, r := c.Response(), c.Request()
	ws, err := tnl.wsu.Upgrade(w, r, nil)
	if err != nil {
		c.Warnf("请求通道连接升级为 websocket 协议时出错：%s", err)
		return nil
	}
	c.Debugf("请求通道连接升级为 websocket 协议成功")

	var mux muxconn.Muxer
	ctx := context.Background()
	conn, proto := ws.NetConn(), ws.Subprotocol()
	switch proto {
	case "vela":
		mux, err = muxconn.NewVela(ctx, conn, nil, true)
	case "yamux":
		mux, err = muxconn.NewYaMUX(ctx, conn, nil, true)
	default:
		mux, err = muxconn.NewSMUX(ctx, conn, nil, true)
	}
	raddr := c.RemoteAddr()
	if err != nil {
		_ = ws.Close()
		c.Warnf("通道 %s 升级多路复用 %s 出错：%s", raddr, proto, err)
		return nil
	}

	c.Infof("通道 %s 升级多路复用 %s 成功", raddr, proto)
	err = tnl.han.ServeMUX(mux)
	_ = mux.Close()
	c.Infof("通道 %s (%s) 断开连接: %v", raddr, proto, err)

	return nil
}
