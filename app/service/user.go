package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/integration/casauth"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"gorm.io/gen/field"
)

type UserService interface {
	// Page 用户分页查询
	Page(ctx context.Context, page param.Pager) (int64, mrequest.UserSummaries)

	Indices(ctx context.Context, indexer param.Indexer) mrequest.UserSummaries

	Delete(ctx context.Context, id int64) error

	Create(ctx context.Context, req *mrequest.UserCreate, cid int64) error

	Sudo(ctx context.Context, req *mrequest.UserSudo) (bool, error)

	Passwd(ctx context.Context, id int64, original string, password string) error

	AccessKey(ctx context.Context, id int64) error

	Authenticate(ctx context.Context, uname, passwd string) (*model.User, error)

	Totp(ctx context.Context, uid int64) error

	Generate(ctx context.Context) error
}

func User(qry *query.Query, digest DigestService, sso casauth.Client, log *slog.Logger) UserService {
	return &userService{
		qry:    qry,
		digest: digest,
		sso:    sso,
		log:    log,
	}
}

type userService struct {
	qry    *query.Query
	digest DigestService
	sso    casauth.Client
	log    *slog.Logger
}

func (biz *userService) Page(ctx context.Context, page param.Pager) (int64, mrequest.UserSummaries) {
	tbl := biz.qry.User
	db := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Username, tbl.Nickname, tbl.Dong, tbl.Enable, tbl.AccessKey)
	if kw := page.Keyword(); kw != "" {
		db.Where(tbl.Username.Like(kw)).
			Or(tbl.Nickname.Like(kw))
	}
	count, err := db.Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	ret := make(mrequest.UserSummaries, 0, page.Size())
	_ = db.Scopes(page.Scope(count)).Scan(&ret)

	return count, ret
}

func (biz *userService) Indices(ctx context.Context, indexer param.Indexer) mrequest.UserSummaries {
	tbl := biz.qry.User
	db := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Username, tbl.Nickname, tbl.Dong, tbl.Enable)
	if kw := indexer.Keyword(); kw != "" {
		db.Where(tbl.Username.Like(kw)).
			Or(tbl.Nickname.Like(kw))
	}

	ret := make(mrequest.UserSummaries, 0, indexer.Size())
	_ = db.Scopes(indexer.Scope).Scan(&ret)
	return ret
}

func (biz *userService) Authenticate(ctx context.Context, uname, passwd string) (*model.User, error) {
	tbl := biz.qry.User
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
	tbl := biz.qry.User
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

func (biz *userService) Create(ctx context.Context, req *mrequest.UserCreate, cid int64) error {
	uname := req.Username
	tbl := biz.qry.User
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
		dat.Password = pwd
	}

	return tbl.WithContext(ctx).Create(dat)
}

func (biz *userService) Sudo(ctx context.Context, req *mrequest.UserSudo) (bool, error) {
	// 查询用户信息
	tbl := biz.qry.User
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
	tbl := biz.qry.User
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

func (biz *userService) AccessKey(ctx context.Context, id int64) error {
	tbl := biz.qry.User
	if count, _ := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Count(); count == 0 {
		return errcode.ErrNotExist
	}

	temp := make([]byte, 32)
	if _, err := rand.Read(temp); err != nil {
		return err
	}
	hec := hex.EncodeToString(temp)
	nano := time.Now().UnixNano()
	sid := strconv.FormatInt(id, 32)
	sat := strconv.FormatInt(nano, 32)

	words := []string{"access", sid, sat, hec}
	ak := strings.Join(words, ".")
	_, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).UpdateColumn(tbl.AccessKey, ak)

	return err
}

func (biz *userService) Totp(ctx context.Context, uid int64) error {
	tbl := biz.qry.User
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(uid)).
		UpdateSimple(tbl.TotpBind.Value(false), tbl.TotpSecret.Value(""))
	return err
}

func (biz *userService) Generate(ctx context.Context) error {
	tbl := biz.qry.User
	dao := tbl.WithContext(ctx)
	cnt, err := dao.Count()
	if err != nil {
		return err
	} else if cnt != 0 {
		biz.log.Info("当前已存在用户，无需生成超级管理员")
		return nil
	}

	buf := make([]byte, 16)
	if _, err = rand.Read(buf); err != nil {
		biz.log.Error("生成初始密码错误", slog.Any("error", err))
		return err
	}
	passwd := hex.EncodeToString(buf)
	const uname = "root"
	hashed, err := biz.digest.Hashed(passwd)
	if err != nil {
		return err
	}

	super := &model.User{
		Username: uname,
		Nickname: "超级管理员",
		Password: hashed,
		Enable:   true,
		Domain:   model.UdLocal,
	}
	if err = dao.Create(super); err != nil {
		biz.log.Error("初始话超级管理员错误", slog.Any("error", err))
		return err
	}
	biz.log.Warn("超级管理员初始化成功",
		slog.String("username", uname),
		slog.String("password", passwd),
	)

	return nil
}
