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
	Buckets(ctx context.Context) []string
	Keys(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.KVData)
	Sweep(ctx context.Context, bucket, key string) error
}

func Shared() SharedService {
	return new(sharedService)
}

type sharedService struct{}

func (svc *sharedService) Buckets(ctx context.Context) []string {
	ret := make([]string, 0, 1024)
	tbl := query.KVData
	_ = tbl.WithContext(ctx).
		Distinct(tbl.Bucket).
		Limit(10000).
		Scan(&ret)

	return ret
}

func (svc *sharedService) Keys(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.KVData) {
	tbl := query.KVData
	db := tbl.WithContext(ctx).
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

func (svc *sharedService) Sweep(ctx context.Context, bucket, key string) error {
	tbl := query.KVData
	if bucket != "" && key != "" {
		_, err := tbl.WithContext(ctx).
			Where(tbl.Bucket.Eq(bucket), tbl.Key.Eq(key)).
			Delete()
		return err
	}

	now := time.Now()
	_, err := tbl.WithContext(ctx).
		Where(tbl.Lifetime.Gt(0), tbl.ExpiredAt.Lt(now)).
		Delete()

	return err
}
