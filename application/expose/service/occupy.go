package service

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/application/expose/response"
)

type Occupy struct {
	qry *query.Query
	log *slog.Logger
}

func NewOccupy(qry *query.Query, log *slog.Logger) *Occupy {
	return &Occupy{
		qry: qry,
		log: log,
	}
}

func (occ *Occupy) Events(ctx context.Context, req *request.OccupyPages) (*response.Pages[response.OccupyStat], error) {
	tbl := occ.qry.Event
	count := tbl.ID.Count()

	dao := tbl.WithContext(ctx).
		Select(tbl.MinionID, tbl.Inet, tbl.FromCode, count.As("count"))

	if id := req.MinionID; id != 0 {
		dao = dao.Where(tbl.MinionID.Eq(id))
	}
	if codes := req.FromCode; len(codes) != 0 {
		dao = dao.Where(tbl.FromCode.In(codes...))
	}
	if occurAt := req.OccurAt; !occurAt.IsZero() {
		dao = dao.Where(tbl.OccurAt.Gte(occurAt))
	}
	dao = dao.Group(tbl.MinionID, tbl.Inet, tbl.FromCode)

	res := response.NewPages[response.OccupyStat](req.PageSize())
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return res, nil
	}

	var records []*response.OccupyStat
	if err = dao.Scopes(res.Scope(cnt)).
		Order(count.Desc()).
		Scan(&records); err != nil {
		return nil, err
	}

	return res.SetRecords(records), nil
}

func (occ *Occupy) Risks(ctx context.Context, req *request.OccupyPages) (*response.Pages[response.OccupyStat], error) {
	tbl := occ.qry.Risk
	count := tbl.ID.Count()

	dao := tbl.WithContext(ctx).
		Select(tbl.MinionID, tbl.Inet, tbl.FromCode, count.As("count"))

	if id := req.MinionID; id != 0 {
		dao = dao.Where(tbl.MinionID.Eq(id))
	}
	if codes := req.FromCode; len(codes) != 0 {
		dao = dao.Where(tbl.FromCode.In(codes...))
	}
	if occurAt := req.OccurAt; !occurAt.IsZero() {
		dao = dao.Where(tbl.OccurAt.Gte(occurAt))
	}
	dao = dao.Group(tbl.MinionID, tbl.Inet, tbl.FromCode)

	res := response.NewPages[response.OccupyStat](req.PageSize())
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return res, nil
	}

	var records []*response.OccupyStat
	if err = dao.Scopes(res.Scope(cnt)).
		Order(count.Desc()).
		Scan(&records); err != nil {
		return nil, err
	}

	return res.SetRecords(records), nil
}
