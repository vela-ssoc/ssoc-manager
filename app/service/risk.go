package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type RiskService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Risk)
	Attack(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.RiskAttack)
	Group(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.NameCount)
	Recent(ctx context.Context, day int) *param.RecentCharts
}

func Risk() RiskService {
	return &riskService{}
}

type riskService struct{}

func (rsk *riskService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Risk) {
	tbl := query.Risk
	db := tbl.WithContext(ctx).
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

func (rsk *riskService) Attack(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.RiskAttack) {
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

func (rsk *riskService) Group(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.NameCount) {
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

func (rsk *riskService) Recent(ctx context.Context, day int) *param.RecentCharts {
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
