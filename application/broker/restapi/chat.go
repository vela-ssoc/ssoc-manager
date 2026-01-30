package restapi

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xgfone/ship/v5"
)

type Chat struct {
	wsu *websocket.Upgrader
}

func NewChat() *Chat {
	return &Chat{
		wsu: &websocket.Upgrader{
			HandshakeTimeout:  time.Minute * 10,
			CheckOrigin:       func(*http.Request) bool { return true },
			EnableCompression: true,
		},
	}
}

func (cht *Chat) BindRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/chat").GET(cht.echo)
	rgb.Route("/download").GET(cht.download)
	return nil
}

func (cht *Chat) echo(c *ship.Context) error {
	w, r := c.Response(), c.Request()
	ws, err := cht.wsu.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	c.Infof("websocket 成功")
	conn := ws.NetConn()
	io.Copy(os.Stdout, conn)

	return nil
}

func (cht *Chat) download(c *ship.Context) error {
	return c.Attachment("D:\\Programs\\Hyper-V\\iso\\Win11_25H2_Chinese_Simplified_x64.iso", "")
}
