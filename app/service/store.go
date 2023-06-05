package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/gen/field"
	"gorm.io/gorm"
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
	Upsert(ctx context.Context, req *param.StoreUpsert) error
	Delete(ctx context.Context, id string) error
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
	tbl := query.Store
	return tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
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

func (biz *storeService) Upsert(ctx context.Context, req *param.StoreUpsert) error {
	id, val := req.ID, req.Value
	cfg := biz.hm[id]
	if cfg != nil {
		if err := cfg.Validate(val); err != nil {
			return err
		}
	}

	tbl := query.Store
	_, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		assigns := []field.AssignExpr{
			tbl.Value.Value(val),
			tbl.Desc.Value(req.Desc),
			tbl.Version.Value(req.Version + 1),
		}
		_, err = tbl.WithContext(ctx).
			Where(tbl.ID.Eq(id), tbl.Version.Eq(req.Version)).
			UpdateColumnSimple(assigns...)
	} else {
		dat := &model.Store{ID: req.ID, Value: req.Value, Desc: req.Desc}
		err = tbl.WithContext(ctx).Create(dat)
	}
	if err != nil || cfg == nil {
		return err
	}
	cfg.Reset()
	if cfg.Shared() {
		biz.pusher.StoreReset(ctx, id)
	}

	return nil
}

func (biz *storeService) Delete(ctx context.Context, id string) error {
	cfg := biz.hm[id]
	tbl := query.Store
	ret, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Delete()
	if err != nil || cfg == nil {
		return err
	}
	if ret.RowsAffected == 0 {
		return errcode.ErrDeleteFailed
	}

	cfg.Reset()
	if cfg.Shared() {
		biz.pusher.StoreReset(ctx, id)
	}

	return nil
}
