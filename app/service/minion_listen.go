package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type MinionListenService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionListen)
}

func MinionListen(qry *query.Query) MinionListenService {
	return &minionListenService{
		qry: qry,
	}
}

type minionListenService struct {
	qry *query.Query
}

func (biz *minionListenService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionListen) {
	tbl := biz.qry.MinionListen
	db := tbl.WithContext(ctx).
		Order(tbl.ID.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if err := db.Count(&count).Error; err != nil || count == 0 {
		return 0, nil
	}

	ret := make([]*model.MinionListen, 0, page.Size())
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
