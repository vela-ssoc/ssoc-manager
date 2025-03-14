package service

import (
	"context"
	"sync"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/ssoc-common-mb/memoize"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
)

func NewAlertServer(qry *query.Query) *AlertServer {
	als := &AlertServer{qry: qry}
	als.mem = memoize.NewTTL2(als.slowLoad, time.Hour)

	return als
}

type AlertServer struct {
	qry   *query.Query
	mutex sync.Mutex
	mem   memoize.TTLCache2[*dong.AlertConfig, error]
}

func (als *AlertServer) Find(ctx context.Context) (*model.AlertServer, error) {
	return als.first(ctx)
}

func (als *AlertServer) Upsert(ctx context.Context, req *mrequest.AlertServerUpsert) error {
	now := time.Now()
	als.mutex.Lock()
	defer als.mutex.Unlock()

	old, _ := als.first(ctx)
	data := &model.AlertServer{
		Mode:      req.Mode,
		Name:      req.Name,
		URL:       req.URL,
		Token:     req.Token,
		Account:   req.Account,
		UpdatedAt: now,
	}
	if old != nil {
		data.ID = old.ID
		data.CreatedAt = old.CreatedAt
	}

	tbl := als.qry.AlertServer
	err := tbl.WithContext(ctx).Save(data)
	if err == nil {
		als.mem.Forget()
	}

	return err
}

func (als *AlertServer) Delete(ctx context.Context) error {
	als.mutex.Lock()
	defer als.mutex.Unlock()

	tbl := als.qry.AlertServer
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Neq(0)). // 跳过全表删除检查。
		Delete()
	if err == nil {
		als.mem.Forget()
	}

	return err
}

func (als *AlertServer) AlertConfigure(ctx context.Context) (*dong.AlertConfig, error) {
	return als.mem.Load(ctx)
}

func (als *AlertServer) slowLoad(ctx context.Context) (*dong.AlertConfig, error) {
	dat, err := als.first(ctx)
	if err != nil {
		return nil, err
	}
	cfg := &dong.AlertConfig{
		SIEM:    dat.Mode == "siem",
		URL:     dat.URL,
		Token:   dat.Token,
		Account: dat.Account,
	}

	return cfg, nil
}

func (als *AlertServer) first(ctx context.Context) (*model.AlertServer, error) {
	tbl := als.qry.AlertServer
	return tbl.WithContext(ctx).
		Order(tbl.CreatedAt.Desc()).
		First()
}
