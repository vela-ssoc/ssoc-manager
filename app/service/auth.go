package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-manager/app/internal/modview"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

// AuthService 认证模块业务层
type AuthService interface {
	Picture(ctx context.Context, uname string) (*param.AuthPicture, error)
	Verify(ctx context.Context, av param.AuthVerify) (factor bool, err error)
	Dong(ctx context.Context, ad param.AuthDong, view modview.LoginDong) error
	Login(ctx context.Context, ab param.AuthLogin) (*model.User, error)
}

func Auth(verify VerifyService, lock LoginLockService, user UserService) AuthService {
	return &authService{
		verify: verify,
		lock:   lock,
		user:   user,
	}
}

type authService struct {
	verify VerifyService
	lock   LoginLockService
	user   UserService
}

func (ath *authService) Picture(ctx context.Context, uname string) (*param.AuthPicture, error) {
	return ath.verify.Picture(ctx, uname)
}

func (ath *authService) Verify(ctx context.Context, av param.AuthVerify) (bool, error) {
	uname, captID := av.Username, av.ID
	points := av.Points.Convert()
	return ath.verify.Verify(ctx, uname, captID, points)
}

func (ath *authService) Dong(ctx context.Context, ad param.AuthDong, view modview.LoginDong) error {
	return ath.verify.DongCode(ctx, ad.Username, ad.CaptchaID, view)
}

func (ath *authService) Login(ctx context.Context, ab param.AuthLogin) (*model.User, error) {
	// 校验验证码
	uname, passwd := ab.Username, ab.Password
	captID, dong := ab.CaptchaID, ab.Code
	if err := ath.verify.Submit(ctx, uname, captID, dong); err != nil {
		return nil, err
	}

	// 检查是否被限制登录
	if ath.lock.Limited(ctx, uname) {
		return nil, errcode.ErrTooManyLoginFailed
	}

	// 开始认证
	user, err := ath.user.Authenticate(ctx, uname, passwd)
	if err != nil {
		ath.lock.Failed(ctx, uname)
		return nil, err
	}
	ath.lock.Passed(ctx, uname)

	return user, nil
}
