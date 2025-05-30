package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/integration/oauth"
	"github.com/vela-ssoc/ssoc-manager/library/totp"
	"github.com/vela-ssoc/ssoc-manager/param/modview"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
)

// AuthService 认证模块业务层
type AuthService interface {
	Picture(ctx context.Context, uname string) (*mrequest.AuthPicture, error)
	Verify(ctx context.Context, av mrequest.AuthVerify) (factor bool, err error)
	Dong(ctx context.Context, ad mrequest.AuthDong, view modview.LoginDong) error
	Login(ctx context.Context, ab mrequest.AuthLogin) (*model.User, error)

	Valid(ctx context.Context, uname, passwd string) (string, bool, error)
	Totp(ctx context.Context, uid string) (*totp.TOTP, error)
	Submit(ctx context.Context, uid, code string) (*model.User, error)
	Oauth(ctx context.Context, req *mrequest.AuthOauth) (*model.User, error)
}

func Auth(qry *query.Query, verify VerifyService, lock LoginLockService, user UserService, oauth oauth.Client) AuthService {
	return &authService{
		qry:    qry,
		verify: verify,
		lock:   lock,
		user:   user,
		oauth:  oauth,
	}
}

type authService struct {
	qry    *query.Query
	verify VerifyService
	lock   LoginLockService
	user   UserService
	oauth  oauth.Client
}

func (svc *authService) Picture(ctx context.Context, uname string) (*mrequest.AuthPicture, error) {
	return svc.verify.Picture(ctx, uname)
}

func (svc *authService) Verify(ctx context.Context, av mrequest.AuthVerify) (bool, error) {
	uname, captID := av.Username, av.ID
	points := av.Points.Convert()
	return svc.verify.Verify(ctx, uname, captID, points)
}

func (svc *authService) Dong(ctx context.Context, ad mrequest.AuthDong, view modview.LoginDong) error {
	return svc.verify.DongCode(ctx, ad.Username, ad.CaptchaID, view)
}

func (svc *authService) Login(ctx context.Context, ab mrequest.AuthLogin) (*model.User, error) {
	// 校验验证码
	uname, passwd := ab.Username, ab.Password
	captID, dong := ab.CaptchaID, ab.Code
	if err := svc.verify.Submit(ctx, uname, captID, dong); err != nil {
		return nil, err
	}

	// 检查是否被限制登录
	if svc.lock.Limited(ctx, uname) {
		return nil, errcode.ErrTooManyLoginFailed
	}

	// 开始认证
	user, err := svc.user.Authenticate(ctx, uname, passwd)
	if err != nil {
		svc.lock.Failed(ctx, uname)
		return nil, err
	}
	svc.lock.Passed(ctx, uname)

	return user, nil
}

func (svc *authService) Valid(ctx context.Context, uname, passwd string) (string, bool, error) {
	if svc.lock.Limited(ctx, uname) {
		return "", false, errcode.ErrTooManyLoginFailed
	}
	user, err := svc.user.Authenticate(ctx, uname, passwd)
	if err != nil {
		svc.lock.Failed(ctx, uname)
		return "", false, err
	}
	svc.lock.Passed(ctx, uname)

	// 生成唯一 UID
	temp := make([]byte, 32)
	_, _ = rand.Read(temp)
	uid := hex.EncodeToString(temp)
	// 保存 uid
	tmp := &model.AuthTemp{
		ID:        user.ID,
		UID:       uid,
		CreatedAt: time.Now(),
	}
	tbl := svc.qry.AuthTemp
	if _, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(user.ID)).Delete(); err != nil {
		return "", false, err
	}
	if err = tbl.WithContext(ctx).Create(tmp); err != nil {
		return "", false, err
	}

	return uid, user.TotpBind, nil
}

func (svc *authService) Totp(ctx context.Context, uid string) (*totp.TOTP, error) {
	now := time.Now()
	tempTbl := svc.qry.AuthTemp
	temp, err := tempTbl.WithContext(ctx).Where(tempTbl.UID.Eq(uid)).First()
	if err != nil || temp.Expired(now, 3*time.Minute) {
		return nil, errcode.ErrUnauthorized
	}
	userTbl := svc.qry.User
	user, err := userTbl.WithContext(ctx).
		Where(userTbl.ID.Eq(temp.ID), userTbl.Enable.Is(true)).
		First()
	if err != nil {
		return nil, errcode.ErrUnauthorized
	}
	if user.TotpBind && user.TotpSecret != "" {
		return nil, errcode.ErrTotpBound
	}

	// 生成一个 totp
	otp := totp.Generate("ssoc", user.Username)
	// 保存 OTP
	_, err = userTbl.WithContext(ctx).
		Where(userTbl.ID.Eq(user.ID)).
		UpdateSimple(userTbl.TotpBind.Value(false), userTbl.TotpSecret.Value(otp.Secret))
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (svc *authService) Submit(ctx context.Context, uid, code string) (*model.User, error) {
	now := time.Now()
	tempTbl := svc.qry.AuthTemp
	temp, err := tempTbl.WithContext(ctx).Where(tempTbl.UID.Eq(uid)).First()
	if err != nil || temp.Expired(now, 3*time.Minute) {
		return nil, errcode.ErrUnauthorized
	}
	userTbl := svc.qry.User
	user, err := userTbl.WithContext(ctx).
		Where(userTbl.ID.Eq(temp.ID), userTbl.Enable.Is(true)).
		First()
	if err != nil {
		return nil, errcode.ErrUnauthorized
	}
	if !totp.Validate(user.TotpSecret, code, false) {
		return nil, errcode.ErrVerifyCode
	}
	if !user.TotpBind {
		_, _ = userTbl.WithContext(ctx).
			Where(userTbl.ID.Eq(user.ID)).
			UpdateColumn(userTbl.TotpBind, true)
	}
	_, _ = tempTbl.WithContext(ctx).Where(tempTbl.UID.Eq(uid)).Delete()

	return user, nil
}

func (svc *authService) Oauth(ctx context.Context, req *mrequest.AuthOauth) (*model.User, error) {
	userinfo, err := svc.oauth.Exchange(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	jobNumber := userinfo.SUB
	tbl := svc.qry.User
	user, err := tbl.WithContext(ctx).Where(tbl.Dong.Eq(jobNumber)).First()
	if err != nil || !user.Enable {
		return nil, errcode.ErrUnauthorized
	}

	return user, nil
}
