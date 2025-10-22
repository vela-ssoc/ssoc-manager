package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/storage/v2"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
)

func NewStartup(qry *query.Query, store storage.Storer, pusher push.Pusher) *Startup {
	return &Startup{
		qry:    qry,
		store:  store,
		pusher: pusher,
	}
}

type Startup struct {
	qry    *query.Query
	store  storage.Storer
	pusher push.Pusher
}

func (stp *Startup) Update(ctx context.Context, req *param.StartupUpdate) error {
	// 查询节点状态
	minionID := req.ID
	monTbl := stp.qry.Minion
	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.Status, monTbl.BrokerID).
		Where(monTbl.ID.Eq(minionID)).
		First()
	if err != nil {
		return err
	}
	if mon.Status == model.MSDelete {
		return errcode.ErrNodeStatus
	}

	data := &model.Startup{
		ID:     minionID,
		Logger: req.Logger,
	}
	tbl := stp.qry.Startup
	dao := tbl.WithContext(ctx)
	err = dao.Where(tbl.ID.Eq(minionID)).Save(data)
	if err == nil {
		stp.pusher.Startup(ctx, mon.BrokerID, minionID)
	}

	return err
}

func (stp *Startup) UpdateFallback(ctx context.Context, req *param.StartupFallbackUpdate) error {
	data := &model.StartupFallback{ID: 1}
	if req != nil {
		data.Logger = req.Logger
	}

	tbl := stp.qry.StartupFallback
	dao := tbl.WithContext(ctx)

	return dao.Save(data)
}

func (stp *Startup) Get(ctx context.Context, minionID int64) (*model.Startup, error) {
	tbl := stp.qry.Startup
	dao := tbl.WithContext(ctx)
	if data, err := dao.Where(tbl.ID.Eq(minionID)).First(); err == nil {
		return data, nil
	}

	fb, err := stp.Fallback(ctx)
	if err != nil {
		return nil, err
	}

	data := &model.Startup{
		ID:        minionID,
		Logger:    fb.Logger,
		CreatedAt: fb.CreatedAt,
		UpdatedAt: fb.UpdatedAt,
	}

	return data, nil
}

func (stp *Startup) Fallback(ctx context.Context) (*model.StartupFallback, error) {
	tbl := stp.qry.StartupFallback
	dao := tbl.WithContext(ctx)
	return dao.Where(tbl.ID.Eq(1)).First()
}
