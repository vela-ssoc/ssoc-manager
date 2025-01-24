package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type AccountService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionAccount)
}

func Account(qry *query.Query) AccountService {
	return &accountService{
		qry: qry,
	}
}

type accountService struct {
	qry *query.Query
}

func (biz *accountService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionAccount) {
	tbl := biz.qry.MinionAccount
	db := tbl.WithContext(ctx).
		Order(tbl.ID.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if err := db.Count(&count).Error; err != nil || count == 0 {
		return 0, nil
	}

	ret := make([]*model.MinionAccount, 0, page.Size())
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
