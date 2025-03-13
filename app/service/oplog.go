package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
)

type OplogService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Oplog)
	Delete(ctx context.Context, scope dynsql.Scope) error
}

func Oplog(qry *query.Query) OplogService {
	return &oplogService{
		qry: qry,
	}
}

type oplogService struct {
	qry *query.Query
}

func (op *oplogService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Oplog) {
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

func (op *oplogService) Delete(ctx context.Context, scope dynsql.Scope) error {
	db := op.qry.Oplog.WithContext(ctx).UnderlyingDB()
	ret := db.Scopes(scope.Where).Delete(&model.Oplog{})
	if ret.Error != nil || ret.RowsAffected != 0 {
		return ret.Error
	}
	return errcode.ErrDeleteFailed
}
