package restapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
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
	req := new(request.TunnelOpen)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	proto := req.ProtocolType()
	w, r := c.Response(), c.Request()
	ws, err := tnl.wsup.Upgrade(w, r, nil)
	if err != nil {
		c.Warnf("请求通道连接升级为 websocket 协议时出错：%s", err)
		return nil
	}
	c.Debugf("请求通道连接升级为 websocket 协议成功")

	var mux muxconn.Muxer
	ctx := context.Background()
	conn := ws.NetConn()
	if proto == "yamux" {
		mux, err = muxconn.NewYaMUX(ctx, conn, nil, true)
	} else {
		mux, err = muxconn.NewSMUX(ctx, conn, nil, true)
	}

	raddr := c.RemoteAddr()
	if err != nil {
		_ = ws.Close()
		c.Warnf("通道 %s 升级多路复用 %s 出错：%s", raddr, proto, err)
		return nil
	}

	c.Infof("通道 %s 升级多路复用 %s 成功", raddr, proto)
	tnl.acpt.AcceptMUX(mux)
	mux.Close()
	c.Infof("通道 %s (%s) 断开连接", raddr, proto)

	return nil
}
