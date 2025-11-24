package mgtapi

import (
	"context"
	"strconv"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/integration/vulnsync"
	"github.com/xgfone/ship/v5"
)

func NewManual(vuln *vulnsync.Synchro) *Manual {
	return &Manual{
		vuln: vuln,
	}
}

type Manual struct {
	vuln *vulnsync.Synchro
}

func (rst *Manual) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/manual/vuln/sync").Data("手动同步漏洞库").PATCH(rst.Sync)
}

func (rst *Manual) Sync(c *ship.Context) error {
	fn := func(sync bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Hour)
		defer cancel()
		_ = rst.vuln.Scan(ctx, sync)
	}

	query := c.Query("sync")
	sync, _ := strconv.ParseBool(query)
	go fn(sync)

	return nil
}
