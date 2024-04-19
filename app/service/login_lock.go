package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
)

type LoginLockService interface {
	// Limited 判断该用户名是被密码错误次数限制
	Limited(ctx context.Context, uname string) bool

	// Failed 密码错误会调用该方法
	Failed(ctx context.Context, uname string)

	// Passed 密码校验成功通过会调用该方法
	Passed(ctx context.Context, uname string)
}

func LoginLock(gap time.Duration, num int) LoginLockService {
	return &loginLockService{
		gap: gap,
		num: num,
	}
}

type loginLockService struct {
	gap time.Duration
	num int
}

func (lck *loginLockService) Limited(ctx context.Context, uname string) bool {
	if lck.gap < time.Second || lck.num <= 0 {
		return false
	}

	afterAt := time.Now().Add(-lck.gap)
	tbl := query.LoginLock
	count, err := tbl.WithContext(ctx).
		Where(tbl.Username.Eq(uname)).
		Where(tbl.CreatedAt.Gte(afterAt)).
		Count()

	return err == nil && int(count) >= lck.num
}

func (lck *loginLockService) Failed(ctx context.Context, uname string) {
	dat := &model.LoginLock{Username: uname}
	_ = query.LoginLock.WithContext(ctx).Create(dat)
}

func (lck *loginLockService) Passed(ctx context.Context, uname string) {
	tbl := query.LoginLock
	_, _ = tbl.WithContext(ctx).
		Where(tbl.Username.Eq(uname)).
		Delete()
}
