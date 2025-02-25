package shipx

import (
	"log/slog"

	"github.com/xgfone/ship/v5"
)

func NotFound(_ *ship.Context) error {
	return ship.ErrNotFound.Newf("资源不存在")
}

func HandleError(c *ship.Context, err error) {
	c.Warnf("ship错误", slog.Any("error", err))
}
