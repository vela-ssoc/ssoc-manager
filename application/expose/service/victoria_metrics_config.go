package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
)

type VictoriaMetricsConfig struct {
	qry *query.Query
	log *slog.Logger
}

func NewVictoriaMetricsConfig(qry *query.Query, log *slog.Logger) *VictoriaMetricsConfig {
	return &VictoriaMetricsConfig{
		qry: qry,
		log: log,
	}
}

func (vmc *VictoriaMetricsConfig) List(ctx context.Context) ([]*model.VictoriaMetricsConfig, error) {
	tbl := vmc.qry.VictoriaMetricsConfig
	dao := tbl.WithContext(ctx)
	return dao.Find()
}

func (vmc *VictoriaMetricsConfig) Create(ctx context.Context, req *request.VictoriaMetricsConfigCreate) error {
	now := time.Now()
	enabled := req.Enabled
	dat := &model.VictoriaMetricsConfig{
		Name:      req.Name,
		Enabled:   enabled,
		Method:    req.Method,
		URL:       req.URL,
		Username:  req.Username,
		Password:  req.Password,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if !enabled {
		tbl := vmc.qry.VictoriaMetricsConfig
		dao := tbl.WithContext(ctx)
		return dao.Create(dat)
	}

	return vmc.qry.Transaction(func(tx *query.Query) error {
		tbl := tx.VictoriaMetricsConfig
		dao := tbl.WithContext(ctx)

		if _, err := dao.Where(tbl.Enabled.Is(true)).
			UpdateColumnSimple(tbl.Enabled.Value(false)); err != nil {
			return err
		}

		return dao.Create(dat)
	})
}

func (vmc *VictoriaMetricsConfig) Update(ctx context.Context, req *request.VictoriaMetricsConfigUpdate) error {
	id, enabled := req.ID, req.Enabled
	tbl := vmc.qry.VictoriaMetricsConfig
	dao := tbl.WithContext(ctx)
	dat, err := dao.Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}

	dat.Name = req.Name
	dat.Enabled = enabled
	dat.Method = req.Method
	dat.URL = req.URL
	dat.Username = req.Username
	dat.Password = req.Password
	if !enabled {
		_, err = dao.Where(tbl.ID.Eq(id)).Updates(dat)
		return err
	}

	return vmc.qry.Transaction(func(tx *query.Query) error {
		tblx := tx.VictoriaMetricsConfig
		daox := tblx.WithContext(ctx)

		if _, err1 := daox.Where(tblx.Enabled.Is(true)).
			UpdateColumnSimple(tblx.Enabled.Value(false)); err1 != nil {
			return err1
		}
		_, err1 := daox.Where(tblx.ID.Eq(id)).Updates(dat)

		return err1
	})
}

func (vmc *VictoriaMetricsConfig) Delete(ctx context.Context, id int64) error {
	tbl := vmc.qry.VictoriaMetricsConfig
	dao := tbl.WithContext(ctx)
	_, err := dao.Where(tbl.ID.Eq(id)).Delete()

	return err
}
