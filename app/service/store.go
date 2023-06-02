package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
)

type StoreConfigurer interface {
	ID() string
	Shared() bool
	Reset()
	Validate([]byte) error
}

type StoreService interface {
	FindID(ctx context.Context, id string) (*model.Store, error)
	Page(ctx context.Context, page param.Pager) (int64, []*model.Store)
}

type storeService struct {
	pusher push.Pusher
	hm     map[string]StoreConfigurer
}

func Store(pusher push.Pusher, ss ...StoreConfigurer) StoreService {
	hm := make(map[string]StoreConfigurer, len(ss))
	for _, s := range ss {
		id := s.ID()
		hm[id] = s
	}
	return &storeService{
		pusher: pusher,
		hm:     hm,
	}
}

func (biz *storeService) FindID(ctx context.Context, id string) (*model.Store, error) {
	return biz.findID(ctx, id)
}

func (biz *storeService) Page(ctx context.Context, page param.Pager) (int64, []*model.Store) {
	tbl := query.Store
	count, err := tbl.WithContext(ctx).Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	dats, _ := tbl.WithContext(ctx).Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *storeService) findID(ctx context.Context, id string) (*model.Store, error) {
	tbl := query.Store
	return tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
}
