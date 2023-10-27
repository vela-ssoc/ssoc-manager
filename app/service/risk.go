package service

import (
	"bytes"
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-common-mb/storage/v2"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type RiskService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Risk)
	Attack(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.RiskAttack)
	Group(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.NameCount)
	Recent(ctx context.Context, day int) *param.RecentCharts
	Delete(ctx context.Context, scope dynsql.Scope) error
	Ignore(ctx context.Context, scope dynsql.Scope) error
	Process(ctx context.Context, scope dynsql.Scope) error
	HTML(ctx context.Context, id int64, secret string) *bytes.Buffer
	Payloads(ctx context.Context, page param.Pager, start, end time.Time, riskType string) (int64, []*param.RiskPayload, error)
}

func Risk(store storage.Storer) RiskService {
	return &riskService{
		store: store,
	}
}

type riskService struct {
	store storage.Storer
}

func (biz *riskService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Risk) {
	tbl := query.Risk
	db := tbl.WithContext(ctx).
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

func (biz *riskService) Attack(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.RiskAttack) {
	db := query.Risk.WithContext(ctx).UnderlyingDB().
		Scopes(scope.Where).
		Group("remote_ip, subject")

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*param.RiskAttack
	db.Select("remote_ip", "subject", "COUNT(*) AS count").
		Order("count DESC").
		Scopes(page.DBScope(count)).
		Find(&dats)

	return count, dats
}

func (biz *riskService) Group(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.NameCount) {
	groupBy := scope.GroupColumn()
	db := query.Risk.WithContext(ctx).UnderlyingDB()

	var count int64
	if db.Scopes(scope.Where).Distinct(groupBy).Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*param.NameCount
	db.Select(groupBy+" AS name", "COUNT(*) AS count").
		Scopes(scope.Where).
		Scopes(scope.GroupBy).
		Order("count DESC").
		Scopes(page.DBScope(count)).
		Find(&dats)

	return count, dats
}

func (biz *riskService) Recent(ctx context.Context, day int) *param.RecentCharts {
	rawSQL := "SELECT a.date, a.risk_type, COUNT(*) AS count " +
		"FROM (SELECT DATE_FORMAT(occur_at, '%m-%d') AS date, risk_type " +
		"FROM risk " +
		"WHERE DATE_FORMAT(occur_at, '%Y-%m-%d') > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), '%Y-%m-%d')) a " +
		"GROUP BY a.date, a.risk_type"

	var temps param.RiskRecentTemps
	query.Risk.WithContext(ctx).
		UnderlyingDB().
		Raw(rawSQL, day).
		Scan(&temps)

	return temps.Charts(day)
}

func (biz *riskService) Delete(ctx context.Context, scope dynsql.Scope) error {
	ret := query.Risk.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where).
		Delete(&model.Risk{})
	if ret.Error != nil || ret.RowsAffected != 0 {
		return ret.Error
	}
	return errcode.ErrDeleteFailed
}

func (biz *riskService) Ignore(ctx context.Context, scope dynsql.Scope) error {
	tbl := query.Risk
	col := tbl.Status.ColumnName().String()
	tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where).
		UpdateColumn(col, model.RSIgnore)
	return nil
}

func (biz *riskService) Process(ctx context.Context, scope dynsql.Scope) error {
	tbl := query.Risk
	col := tbl.Status.ColumnName().String()
	tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where).
		UpdateColumn(col, model.RSProcessed)
	return nil
}

func (biz *riskService) HTML(ctx context.Context, id int64, secret string) *bytes.Buffer {
	tbl := query.Risk
	rsk, _ := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id), tbl.Secret.Eq(secret), tbl.SendAlert.Is(true)).
		First()
	if rsk == nil {
		rsk = new(model.Risk)
	}

	return biz.store.RiskHTML(ctx, rsk)
}

func (biz *riskService) Payloads(ctx context.Context, page param.Pager, start, end time.Time, riskType string) (int64, []*param.RiskPayload, error) {
	tbl := query.Risk
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

	var ret []*param.RiskPayload
	err = dao.Scopes(page.Scope(count)).UnderlyingDB().
		Select("DISTINCT(payload)", "id", "occur_at").
		Find(&ret).Error

	return count, ret, err
}
