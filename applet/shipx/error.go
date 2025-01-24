package shipx

import (
	"log/slog"

	"github.com/xgfone/ship/v5"
)

func NotFound(*ship.Context) error {
	return ship.ErrNotFound.Newf("资源未找到")
}

func HandleError(c *ship.Context, err error) {
	c.Warnf("ship错误", slog.Any("error", err))
}
