package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type UserService interface {
	// Page 用户分页查询
	Page(ctx context.Context, page param.Pager) (int64, param.UserSummaries)

	Indices(ctx context.Context, indexer param.Indexer) param.UserSummaries

	Authenticate(ctx context.Context, uname, passwd string) (*model.User, error)
}

func User(digest DigestService) UserService {
	return &userService{
		digest: digest,
	}
}

type userService struct {
	digest DigestService
}

func (usr *userService) Page(ctx context.Context, page param.Pager) (int64, param.UserSummaries) {
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

func (usr *userService) Indices(ctx context.Context, indexer param.Indexer) param.UserSummaries {
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

func (usr *userService) Authenticate(ctx context.Context, uname, passwd string) (*model.User, error) {
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
		if !usr.digest.Compare(user.Password, passwd) {
			return nil, errcode.ErrPassword
		}
		return user, nil
	}

	// sso 认证的账户

	return nil, err
}
