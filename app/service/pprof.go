package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/internal/prof"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
)

func NewPprof(qry *query.Query, dir string, pusher push.Pusher) *Pprof {
	return &Pprof{
		qry:    qry,
		dir:    dir,
		pusher: pusher,
	}
}

type Pprof struct {
	qry    *query.Query
	dir    string
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
	_, _ = rand.Read(buf)
	name := fmt.Sprintf("%d-%d-%x", nano, mid, buf)
	dest := filepath.Join(svc.dir, name)

	if err = svc.pusher.SavePprof(ctx, bid, mid, second, dest, "profile"); err != nil {
		return "", err
	}

	return name, nil
}

func (svc *Pprof) View(_ context.Context, name string) (http.Handler, error) {
	name = filepath.Join("/", name)
	name = filepath.Join(svc.dir, name)
	return prof.New(name)
}

func (svc *Pprof) Dump(ctx context.Context, req *param.PprofDump) (string, error) {
	id := req.ID
	tbl := svc.qry.Minion
	dao := tbl.WithContext(ctx)
	mon, err := dao.Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return "", err
	}

	status := mon.Status
	if status == model.MSOffline || status == model.MSDelete {
		return "", errcode.ErrNodeStatus
	}
	bid, mid := mon.BrokerID, mon.ID

	nano := time.Now().UnixNano()
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	name := fmt.Sprintf("%d-%d-%x", nano, mid, buf)
	dest := filepath.Join(svc.dir, name)

	second := req.Second
	if err = svc.pusher.SavePprof(ctx, bid, mid, second, dest, req.Type); err != nil {
		return "", err
	}

	return name, nil
}
