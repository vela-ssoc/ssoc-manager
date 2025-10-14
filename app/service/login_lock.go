package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
)

func NewLoginLock(qry *query.Query, gap time.Duration, num int) *LoginLock {
	return &LoginLock{
		qry: qry,
		gap: gap,
		num: num,
	}
}

type LoginLock struct {
	qry *query.Query
	gap time.Duration
	num int
}

func (lck *LoginLock) Limited(ctx context.Context, uname string) bool {
	if lck.gap < time.Second || lck.num <= 0 {
		return false
	}

	afterAt := time.Now().Add(-lck.gap)
	tbl := lck.qry.LoginLock
	count, err := tbl.WithContext(ctx).
		Where(tbl.Username.Eq(uname)).
		Where(tbl.CreatedAt.Gte(afterAt)).
		Count()

	return err == nil && int(count) >= lck.num
}

func (lck *LoginLock) Failed(ctx context.Context, uname string) {
	dat := &model.LoginLock{Username: uname}
	_ = lck.qry.LoginLock.WithContext(ctx).Create(dat)
}

func (lck *LoginLock) Passed(ctx context.Context, uname string) {
	tbl := lck.qry.LoginLock
	_, _ = tbl.WithContext(ctx).
		Where(tbl.Username.Eq(uname)).
		Delete()
}
