package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewMinionListen(qry *query.Query) *MinionListen {
	return &MinionListen{
		qry: qry,
	}
}

type MinionListen struct {
	qry *query.Query
}

func (biz *MinionListen) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionListen) {
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
