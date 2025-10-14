package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/app/internal/prof"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
)

func NewPprof(qry *query.Query, dir string, pusher push.Pusher) *Pprof {
	nano := time.Now().UnixNano()
	random := rand.New(rand.NewSource(nano))
	return &Pprof{
		qry:    qry,
		dir:    dir,
		random: random,
		pusher: pusher,
	}
}

type Pprof struct {
	qry    *query.Query
	dir    string
	random *rand.Rand
	pusher push.Pusher
}

func (svc *Pprof) Load(ctx context.Context, node string, second int) (string, error) {
	id, _ := strconv.ParseInt(node, 10, 64)

	tbl := svc.qry.Minion
	dao := tbl.WithContext(ctx).Where(tbl.Inet.Eq(node))
	if id != 0 {
		dao.Or(tbl.ID.Eq(id))
	}
	mon, err := dao.First()
	if err != nil {
		return "", err
	}

	status := mon.Status
	if status == model.MSOffline || status == model.MSDelete {
		return "", errcode.ErrNodeStatus
	}
	bid, mid := mon.BrokerID, mon.ID

	nano := time.Now().UnixNano()
	buf := make([]byte, 16)
	svc.random.Read(buf)
	name := fmt.Sprintf("%d-%d-%x", nano, mid, buf)
	dest := filepath.Join(svc.dir, name)

	if err = svc.pusher.SavePprof(ctx, bid, mid, second, dest); err != nil {
		return "", err
	}

	return name, nil
}

func (svc *Pprof) View(_ context.Context, name string) (http.Handler, error) {
	name = filepath.Join("/", name)
	name = filepath.Join(svc.dir, name)
	return prof.New(name)
}
