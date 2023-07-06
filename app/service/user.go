package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/integration/ssoauth"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/gen/field"
)

type UserService interface {
	// Page 用户分页查询
	Page(ctx context.Context, page param.Pager) (int64, param.UserSummaries)

	Indices(ctx context.Context, indexer param.Indexer) param.UserSummaries

	Delete(ctx context.Context, id int64) error

	Create(ctx context.Context, req *param.UserCreate, cid int64) error

	Sudo(ctx context.Context, req *param.UserSudo) (bool, error)

	Passwd(ctx context.Context, id int64, original string, password string) error

	Authenticate(ctx context.Context, uname, passwd string) (*model.User, error)
}

func User(digest DigestService, sso ssoauth.Client) UserService {
	return &userService{
		digest: digest,
		sso:    sso,
	}
}

type userService struct {
	digest DigestService
	sso    ssoauth.Client
}

func (biz *userService) Page(ctx context.Context, page param.Pager) (int64, param.UserSummaries) {
	tbl := query.User
	db := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Username, tbl.Nickname, tbl.Dong, tbl.Enable)
	if kw := page.Keyword(); kw != "" {
		db.Where(tbl.Username.Like(kw)).
			Or(tbl.Nickname.Like(kw))
	}
	count, err := db.Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	ret := make(param.UserSummaries, 0, page.Size())
	_ = db.Scopes(page.Scope(count)).Scan(&ret)

	return count, ret
}

func (biz *userService) Indices(ctx context.Context, indexer param.Indexer) param.UserSummaries {
	tbl := query.User
	db := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Username, tbl.Nickname, tbl.Dong, tbl.Enable)
	if kw := indexer.Keyword(); kw != "" {
		db.Where(tbl.Username.Like(kw)).
			Or(tbl.Nickname.Like(kw))
	}

	ret := make(param.UserSummaries, 0, indexer.Size())
	_ = db.Scopes(indexer.Scope).Scan(&ret)
	return ret
}

func (biz *userService) Authenticate(ctx context.Context, uname, passwd string) (*model.User, error) {
	tbl := query.User
	user, err := tbl.WithContext(ctx).
		Where(tbl.Username.Eq(uname)).
		Where(tbl.Enable.Is(true)).
		First()
	if err != nil {
		return nil, err
	}
	// 本地账户
	if user.IsLocal() {
		if !biz.digest.Compare(user.Password, passwd) {
			return nil, errcode.ErrPassword
		}
		return user, nil
	}

	// sso 认证的账户
	if err = biz.sso.Auth(ctx, uname, passwd); err != nil {
		return nil, err
	}

	return user, nil
}

func (biz *userService) Delete(ctx context.Context, id int64) error {
	tbl := query.User
	ret, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Delete()
	if err != nil || ret.Error == nil {
		return err
	}
	if ret.RowsAffected == 0 {
		return errcode.ErrDeleteFailed
	}
	return nil
}

func (biz *userService) Create(ctx context.Context, req *param.UserCreate, cid int64) error {
	uname := req.Username
	tbl := query.User
	if count, _ := tbl.WithContext(ctx).
		Where(tbl.Username.Eq(uname)).
		Count(); count != 0 {
		return errcode.FmtErrNameExist.Fmt(uname)
	}

	dat := &model.User{
		Username: uname,
		Nickname: req.Nickname,
		Enable:   req.Enable,
		Domain:   req.Domain,
	}
	// OA 用户名与咚咚号同名
	if req.Domain == model.UdOA {
		dat.Dong = req.Username
	} else {
		pwd, err := biz.digest.Hashed(req.Password)
		if err != nil {
			return errcode.ErrPassword
		}
		req.Password = pwd
	}

	return tbl.WithContext(ctx).Create(dat)
}

func (biz *userService) Sudo(ctx context.Context, req *param.UserSudo) (bool, error) {
	// 查询用户信息
	tbl := query.User
	user, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(req.ID)).First()
	if err != nil {
		return false, err
	}
	assigns := []field.AssignExpr{
		tbl.Nickname.Value(req.Nickname),
		tbl.Enable.Value(req.Enable),
	}

	logout := user.Enable != req.Enable
	if user.Domain == model.UdLocal && req.Password != "" {
		pwd, exx := biz.digest.Hashed(req.Password)
		if exx != nil {
			return false, exx
		}
		logout = true
		assigns = append(assigns, tbl.Password.Value(pwd))
	}

	_, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		UpdateColumnSimple(assigns...)

	return logout, err
}

func (biz *userService) Passwd(ctx context.Context, id int64, original string, password string) error {
	tbl := query.User
	user, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}
	if user.Domain != model.UdLocal {
		return errcode.ErrOperateFailed
	}

	if !biz.digest.Compare(user.Password, original) {
		return errcode.ErrPassword
	}
	pwd, err := biz.digest.Hashed(password)
	if err != nil {
		return errcode.ErrPassword
	}

	_, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Update(tbl.Password, pwd)

	return err
}
