package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewProcess(qry *query.Query) *Process {
	return &Process{
		qry: qry,
	}
}

type Process struct {
	qry *query.Query
}

func (biz *Process) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionProcess) {
	tbl := biz.qry.MinionProcess
	db := tbl.WithContext(ctx).
		Order(tbl.ID.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if err := db.Count(&count).Error; err != nil || count == 0 {
		return 0, nil
	}

	ret := make([]*model.MinionProcess, 0, page.Size())
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
