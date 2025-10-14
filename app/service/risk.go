package service

import (
	"bytes"
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-common-mb/param/response"
	"github.com/vela-ssoc/ssoc-common-mb/storage/v2"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
)

func NewRisk(qry *query.Query, store storage.Storer) *Risk {
	return &Risk{
		qry:   qry,
		store: store,
	}
}

type Risk struct {
	qry   *query.Query
	store storage.Storer
}

func (biz *Risk) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Risk) {
	tbl := biz.qry.Risk
	db := tbl.WithContext(ctx).
		Where(tbl.Status.Neq(uint8(model.RSIgnore))).
		Order(tbl.ID.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret []*model.Risk
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}

func (biz *Risk) Attack(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*mrequest.RiskAttack) {
	db := biz.qry.Risk.WithContext(ctx).UnderlyingDB().
		Scopes(scope.Where).
		Group("remote_ip, subject")

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*mrequest.RiskAttack
	db.Select("remote_ip", "subject", "COUNT(*) AS count").
		Order("count DESC").
		Scopes(page.DBScope(count)).
		Find(&dats)

	return count, dats
}

func (biz *Risk) Group(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, request.NameCounts) {
	groupBy := scope.GroupColumn()
	db := biz.qry.Risk.WithContext(ctx).UnderlyingDB()

	var count int64
	if db.Scopes(scope.Where).Distinct(groupBy).Count(&count); count == 0 {
		return 0, nil
	}

	var dats request.NameCounts
	db.Select(groupBy+" AS name", "COUNT(*) AS count").
		Scopes(scope.Where).
		Scopes(scope.GroupBy).
		Order("count DESC").
		Scopes(page.DBScope(count)).
		Find(&dats)

	return count, dats
}

func (biz *Risk) Recent(ctx context.Context, day int) *mrequest.RecentCharts {
	rawSQL := "SELECT a.date, a.risk_type, COUNT(*) AS count " +
		"FROM (SELECT DATE_FORMAT(occur_at, '%m-%d') AS date, risk_type " +
		"FROM risk " +
		"WHERE DATE_FORMAT(occur_at, '%Y-%m-%d') > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), '%Y-%m-%d')) a " +
		"GROUP BY a.date, a.risk_type"

	var temps mrequest.RiskRecentTemps
	biz.qry.Risk.WithContext(ctx).
		UnderlyingDB().
		Raw(rawSQL, day).
		Scan(&temps)

	return temps.Charts(day)
}

func (biz *Risk) Delete(ctx context.Context, scope dynsql.Scope) error {
	ret := biz.qry.Risk.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where).
		Delete(&model.Risk{})
	if ret.Error != nil || ret.RowsAffected != 0 {
		return ret.Error
	}
	return errcode.ErrDeleteFailed
}

func (biz *Risk) Ignore(ctx context.Context, scope dynsql.Scope) error {
	tbl := biz.qry.Risk
	col := tbl.Status.ColumnName().String()
	tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where).
		UpdateColumn(col, model.RSIgnore)
	return nil
}

func (biz *Risk) Process(ctx context.Context, scope dynsql.Scope) error {
	tbl := biz.qry.Risk
	col := tbl.Status.ColumnName().String()
	tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where).
		UpdateColumn(col, model.RSProcessed)
	return nil
}

func (biz *Risk) HTML(ctx context.Context, id int64, secret string) *bytes.Buffer {
	tbl := biz.qry.Risk
	rsk, _ := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id), tbl.Secret.Eq(secret), tbl.SendAlert.Is(true)).
		First()
	if rsk == nil {
		rsk = new(model.Risk)
	}

	return biz.store.RiskHTML(ctx, rsk)
}

func (biz *Risk) PayloadsBAK(ctx context.Context, page param.Pager, start, end time.Time, riskType string) (int64, []*mrequest.RiskPayload, error) {
	tbl := biz.qry.Risk
	dao := tbl.WithContext(ctx).
		Distinct(tbl.Payload).
		Where(tbl.OccurAt.Between(start, end)).
		Order(tbl.ID.Desc())
	if riskType != "" {
		dao.Where(tbl.RiskType.Eq(riskType))
	}

	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil, err
	}

	var ret []*mrequest.RiskPayload
	err = dao.Scopes(page.Scope(count)).UnderlyingDB().
		Select("DISTINCT(payload)", "id", "occur_at").
		Find(&ret).Error

	return count, ret, err
}

func (biz *Risk) Payloads(ctx context.Context, page request.Pages, start, end time.Time, riskType string) (*response.Pages[*mrequest.RiskPayload], error) {
	tbl := biz.qry.Risk
	dao := tbl.WithContext(ctx).
		Distinct(tbl.Payload).
		Where(tbl.OccurAt.Between(start, end)).
		Order(tbl.ID.Desc())
	if riskType != "" {
		dao.Where(tbl.RiskType.Eq(riskType))
	}

	pages := response.NewPages[*mrequest.RiskPayload](page.PageSize())
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return pages.Empty(), nil
	}

	records := make([]*mrequest.RiskPayload, 0, 10)
	if err = dao.Scopes(pages.FP(cnt)).
		Select(tbl.Payload.Distinct(), tbl.ID, tbl.OccurAt).
		Scan(&records); err != nil {
		return nil, err
	}

	return pages.SetRecords(records), nil
}
