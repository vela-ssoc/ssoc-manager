package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/storage/v2"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
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
	FindJSON(ctx context.Context, id string, v any) error
	Page(ctx context.Context, page param.Pager) (int64, []*model.Store)
	Upsert(ctx context.Context, req *param.StoreUpsert) error
	Delete(ctx context.Context, id string) error
}

type storeService struct {
	qry    *query.Query
	pusher push.Pusher
	store  storage.Storer
}

func Store(qry *query.Query, pusher push.Pusher, store storage.Storer) StoreService {
	return &storeService{
		qry:    qry,
		pusher: pusher,
		store:  store,
	}
}

func (biz *storeService) FindID(ctx context.Context, id string) (*model.Store, error) {
	tbl := biz.qry.Store
	return tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
}

func (biz *storeService) FindJSON(ctx context.Context, id string, v any) error {
	dat, err := biz.FindID(ctx, id)
	if err != nil {
		return err
	}
	return json.Unmarshal(dat.Value, v)
}

func (biz *storeService) Page(ctx context.Context, page param.Pager) (int64, []*model.Store) {
	tbl := biz.qry.Store
	count, err := tbl.WithContext(ctx).Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	dats, _ := tbl.WithContext(ctx).Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *storeService) Upsert(ctx context.Context, req *param.StoreUpsert) error {
	id, val := req.ID, req.Value
	if err := biz.store.Validate(id, val); err != nil {
		return err
	}
	//if biz.store.Validate(id, val) {
	//	return errcode.ErrInvalidData
	//}

	tbl := biz.qry.Store
	_, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		dat := &model.Store{ID: req.ID, Value: req.Value, Desc: req.Desc}
		err = tbl.WithContext(ctx).Create(dat)
	} else {
		assigns := []field.AssignExpr{
			tbl.Value.Value(val),
			tbl.Desc.Value(req.Desc),
			tbl.Version.Value(req.Version + 1),
		}
		_, err = tbl.WithContext(ctx).
			Where(tbl.ID.Eq(id), tbl.Version.Eq(req.Version)).
			UpdateColumnSimple(assigns...)
	}
	if err != nil {
		return err
	}
	if biz.store.Forget(id) {
		biz.pusher.StoreReset(ctx, id)
	}

	return nil
}

func (biz *storeService) Delete(ctx context.Context, id string) error {
	tbl := biz.qry.Store
	ret, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Delete()
	if err != nil {
		return err
	}
	if ret.RowsAffected == 0 {
		return errcode.ErrDeleteFailed
	}

	if biz.store.Forget(id) {
		biz.pusher.StoreReset(ctx, id)
	}

	return nil
}
