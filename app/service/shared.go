package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type SharedService interface {
	Keys(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.KVData)
	Sweep(ctx context.Context) error
}

func Shared() SharedService {
	return new(sharedService)
}

type sharedService struct{}

func (svc *sharedService) Keys(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.KVData) {
	tbl := query.KVData
	db := tbl.WithContext(ctx).
		Omit(tbl.Value).
		Order(tbl.UpdatedAt.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.KVData
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (svc *sharedService) Sweep(ctx context.Context) error {
	now := time.Now()
	tbl := query.KVData
	_, err := tbl.WithContext(ctx).
		Where(tbl.Lifetime.Neq(0), tbl.ExpiredAt.Lt(now)).
		Delete()

	return err
}
