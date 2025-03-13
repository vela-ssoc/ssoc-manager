package service

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/memoize"
)

func NewSIEMServer(qry *query.Query) *SIEMServer {
	sim := &SIEMServer{qry: qry}
	sim.mem = memoize.NewTTL2(sim.slowLoad, time.Hour)

	return sim
}

type SIEMServer struct {
	qry   *query.Query
	mutex sync.Mutex
	mem   memoize.TTLCache2[http.Handler, error]
}

func (sim *SIEMServer) SIEMHandler(ctx context.Context) (http.Handler, error) {
	return sim.mem.Load(ctx)
}

func (sim *SIEMServer) Find(ctx context.Context) (*model.SIEMServer, error) {
	return sim.first(ctx)
}

func (sim *SIEMServer) Upsert(ctx context.Context, req *param.SIEMServerUpsert) error {
	now := time.Now()
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	old, _ := sim.first(ctx)
	data := &model.SIEMServer{
		Name:      req.Name,
		URL:       req.URL,
		Token:     req.Token,
		UpdatedAt: now,
	}
	if old != nil {
		data.ID = old.ID
		data.CreatedAt = old.CreatedAt
	}

	tbl := sim.qry.SIEMServer
	err := tbl.WithContext(ctx).Save(data)
	if err == nil {
		sim.mem.Forget()
	}

	return err
}

func (sim *SIEMServer) Delete(ctx context.Context) error {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	tbl := sim.qry.SIEMServer
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Neq(0)). // 跳过全表删除检查。
		Delete()
	if err == nil {
		sim.mem.Forget()
	}

	return err
}

func (sim *SIEMServer) slowLoad(ctx context.Context) (http.Handler, error) {
	dat, err := sim.first(ctx)
	if err != nil {
		return nil, err
	}
	pu, err := url.Parse(dat.URL)
	if err != nil {
		return nil, err
	}
	px := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(pu)
			pr.Out.Header.Set("Authorization", dat.Token)
		},
	}

	return px, nil
}

func (sim *SIEMServer) first(ctx context.Context) (*model.SIEMServer, error) {
	tbl := sim.qry.SIEMServer
	return tbl.WithContext(ctx).
		Order(tbl.CreatedAt.Desc()).
		First()
}
