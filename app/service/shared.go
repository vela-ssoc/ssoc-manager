package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"gorm.io/gen"
)

func NewShared(qry *query.Query) *Shared {
	return &Shared{qry: qry}
}

type Shared struct {
	qry *query.Query
}

func (svc *Shared) Buckets(ctx context.Context) []string {
	ret := make([]string, 0, 1024)
	tbl := svc.qry.KVData
	_ = tbl.WithContext(ctx).
		Distinct(tbl.Bucket).
		Limit(10000).
		Scan(&ret)

	return ret
}

func (svc *Shared) Keys(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.KVData) {
	tbl := svc.qry.KVData
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

// Sweep 清除 kv 数据。
func (svc *Shared) Sweep(ctx context.Context, bucket, key string) error {
	now := time.Now()
	tbl := svc.qry.KVData
	if bucket == "" { // 仅清理过期的数据。
		_, err := tbl.WithContext(ctx).
			Where(tbl.Lifetime.Gt(0), tbl.ExpiredAt.Lt(now)).
			Delete()
		return err
	}

	auditTbl := svc.qry.KVAudit
	dataCond := []gen.Condition{tbl.Bucket.Eq(bucket)}
	auditCond := []gen.Condition{auditTbl.Bucket.Eq(bucket)}
	if key != "" {
		dataCond = append(dataCond, tbl.Key.Eq(key))
		auditCond = append(auditCond, auditTbl.Key.Eq(key))
	}

	_, err := tbl.WithContext(ctx).Where(dataCond...).Delete()
	if err == nil {
		_, _ = auditTbl.WithContext(ctx).Where(auditCond...).Delete()
	}

	return err
}

func (svc *Shared) Audits(ctx context.Context, page param.Pager, bucket, key string) (int64, []*model.KVAudit) {
	tbl := svc.qry.KVAudit
	stmt := tbl.WithContext(ctx).
		Where(tbl.Bucket.Eq(bucket), tbl.Key.Eq(key))
	count, _ := stmt.Count()
	if count <= 0 {
		return 0, nil
	}

	dats, _ := stmt.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (svc *Shared) Update(ctx context.Context, req *param.SharedUpdate) error {
	tbl := svc.qry.KVData
	bucket, key := req.Bucket, req.Key

	now := time.Now()
	if dat, err := tbl.WithContext(ctx).
		Where(tbl.Bucket.Eq(bucket), tbl.Key.Eq(key)).
		First(); err != nil {
		return err
	} else if dat.Expired(now) {
		return errcode.ErrNotExist
	}

	_, err := tbl.WithContext(ctx).
		Where(tbl.Bucket.Eq(bucket), tbl.Key.Eq(key)).
		UpdateSimple(tbl.Value.Value(req.Value))

	return err
}
