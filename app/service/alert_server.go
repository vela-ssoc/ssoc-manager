package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
)

func NewAlertServer(qry *query.Query) *AlertServer {
	return &AlertServer{qry: qry}
}

type AlertServer struct {
	qry *query.Query
}

func (als *AlertServer) Find(ctx context.Context) (*model.AlertServer, error) {
	return als.slowQuery(ctx)
}

func (als *AlertServer) Upsert(ctx context.Context, req *mrequest.AlertServerUpsert) error {
	now := time.Now()

	old, _ := als.slowQuery(ctx)
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

	return tbl.WithContext(ctx).Save(data)
}

func (als *AlertServer) Delete(ctx context.Context) error {
	tbl := als.qry.AlertServer
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Neq(0)). // 跳过全表删除检查。
		Delete()

	return err
}

func (als *AlertServer) AlertConfigure(ctx context.Context) (*dong.AlertConfig, error) {
	dat, err := als.slowQuery(ctx)
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

func (als *AlertServer) slowQuery(ctx context.Context) (*model.AlertServer, error) {
	tbl := als.qry.AlertServer
	return tbl.WithContext(ctx).
		Order(tbl.UpdatedAt.Desc()).
		First()
}
