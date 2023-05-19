package middle

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/xgfone/ship/v5"
)

type OplogSaver interface {
	Save(context.Context, *model.Oplog) error
}

func Oplog(save OplogSaver) ship.Middleware {
	return nil
}

type oplog struct {
	save OplogSaver
}

func (op *oplog) middleFunc(h ship.Handler) ship.Handler {
	hs := &handleShip{han: h}
	return hs.serveShip
}

type handleShip struct {
	han  ship.Handler
	save OplogSaver
}

func (h *handleShip) serveShip(c *ship.Context) error {
	r := c.Request()
	ctx := r.Context()

	err := h.han(c)
	_ = h.save.Save(ctx, nil)

	return err
}
