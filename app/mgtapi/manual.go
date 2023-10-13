package mgtapi

import (
	"context"
	"strconv"
	"time"

	"github.com/vela-ssoc/vela-common-mb/integration/vulnsync"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/xgfone/ship/v5"
)

func Manual(vuln *vulnsync.Synchro) route.Router {
	return &manualREST{
		vuln: vuln,
	}
}

type manualREST struct {
	vuln *vulnsync.Synchro
}

func (rst *manualREST) Route(anon, bearer, basic *ship.RouteGroupBuilder) {
	bearer.Route("/manual/vuln/sync").Data("手动同步漏洞库").PATCH(rst.Sync)
}

func (rst *manualREST) Sync(c *ship.Context) error {
	fn := func(sync bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Hour)
		defer cancel()
		_ = rst.vuln.Scan(ctx, false)
	}

	query := c.Query("sync")
	sync, _ := strconv.ParseBool(query)
	go fn(sync)

	return nil
}
