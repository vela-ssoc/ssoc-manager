package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

type MinionAccountService interface {
	Page(ctx context.Context, page param.Pager, mid int64, name string) (int64, []*model.MinionAccount)
}

func MinionAccount(qry *query.Query) MinionAccountService {
	return &minionAccountService{
		qry: qry,
	}
}

type minionAccountService struct {
	qry *query.Query
}

func (biz *minionAccountService) Page(ctx context.Context, page param.Pager, mid int64, name string) (int64, []*model.MinionAccount) {
	tbl := biz.qry.MinionAccount
	dao := tbl.WithContext(ctx).
		Order(tbl.ID.Desc())
	if mid != 0 {
		dao.Where(tbl.MinionID.Eq(mid))
	}
	if name != "" {
		like := "%" + name + "%"
		dao.Where(tbl.Name.Like(like))
	}

	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}
