package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/errcode"
)

func NewOplog(qry *query.Query) *Oplog {
	return &Oplog{
		qry: qry,
	}
}

type Oplog struct {
	qry *query.Query
}

func (op *Oplog) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Oplog) {
	tbl := op.qry.Oplog
	db := tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if err := db.Count(&count).Error; err != nil || count == 0 {
		return 0, nil
	}

	ret := make([]*model.Oplog, 0, page.Size())
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}

func (op *Oplog) Delete(ctx context.Context, scope dynsql.Scope) error {
	db := op.qry.Oplog.WithContext(ctx).UnderlyingDB()
	ret := db.Scopes(scope.Where).Delete(&model.Oplog{})
	if ret.Error != nil || ret.RowsAffected != 0 {
		return ret.Error
	}
	return errcode.ErrDeleteFailed
}
