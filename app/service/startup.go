package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/storage/v2"
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

func (biz *Startup) Detail(ctx context.Context, id int64) (*model.Startup, error) {
	tbl := biz.qry.Startup
	dat, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err == nil {
		return dat, nil
	}
	// 查询全局配置
	ret, exx := biz.store.Startup(ctx)
	if exx != nil {
		return nil, exx
	}
	ret.ID = id
	ret.Node.ID = id

	return ret, nil
}

func (biz *Startup) Update(ctx context.Context, req *model.Startup) error {
	// 查询节点状态
	monTbl := biz.qry.Minion
	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.Status, monTbl.BrokerID).
		Where(monTbl.ID.Eq(req.ID)).
		First()
	if err != nil {
		return err
	}
	if mon.Status == model.MSDelete {
		return errcode.ErrNodeStatus
	}

	// 更新 startup
	tbl := biz.qry.Startup
	err = tbl.WithContext(ctx).Where(tbl.ID.Eq(req.ID)).Save(req)
	if err == nil {
		biz.pusher.Startup(ctx, mon.BrokerID, req.ID)
	}

	return err
}
