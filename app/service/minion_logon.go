package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type MinionLogonService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionLogon)
	Attack(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.MinionLogonAttack)
	Recent(ctx context.Context, days int) param.MinionRecent
	History(ctx context.Context, page param.Pager, mid int64, name string) (int64, []*model.MinionLogon)
	Ignore(ctx context.Context, id int64) error
}

func MinionLogon() MinionLogonService {
	return &minionLogonService{}
}

type minionLogonService struct{}

func (biz *minionLogonService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionLogon) {
	tbl := query.MinionLogon
	db := tbl.WithContext(ctx).
		Where(tbl.Ignore.Is(false)).
		UnderlyingDB().
		Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.MinionLogon
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (biz *minionLogonService) Attack(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.MinionLogonAttack) {
	tbl := query.MinionLogon
	db := tbl.WithContext(ctx).
		Where(tbl.Ignore.Is(false)).
		Group(tbl.Addr, tbl.Msg).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*param.MinionLogonAttack
	db.Select("addr", "msg", "count(*) AS count").
		Order("count DESC").
		Scopes(page.DBScope(count)).
		Scan(&dats)

	return count, dats
}

func (biz *minionLogonService) Recent(ctx context.Context, days int) param.MinionRecent {
	rawSQL := "SELECT a.date, a.msg, COUNT(*) AS count " +
		"FROM (SELECT DATE_FORMAT(logon_at, '%m-%d') AS date, msg " +
		"      FROM minion_logon " +
		"      WHERE DATE_FORMAT(logon_at, '%Y-%m-%d') > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), '%Y-%m-%d')) a " +
		" GROUP BY a.date, a.msg"

	var temps param.MinionRecentTemps
	query.MinionLogon.
		WithContext(ctx).
		UnderlyingDB().
		Raw(rawSQL, days).
		Scan(&temps)

	res := temps.Format(days)

	return res
}

func (biz *minionLogonService) History(ctx context.Context, page param.Pager, mid int64, name string) (int64, []*model.MinionLogon) {
	tbl := query.MinionLogon
	dao := tbl.WithContext(ctx)
	if mid != 0 {
		dao = dao.Where(tbl.MinionID.Eq(mid))
	}
	if name != "" {
		like := "%" + name + "%"
		dao = dao.Where(tbl.User.Like(like))
	}

	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *minionLogonService) Ignore(ctx context.Context, id int64) error {
	tbl := query.MinionLogon
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id), tbl.Ignore.Is(false)).
		UpdateSimple(tbl.Ignore.Value(true))

	return err
}
