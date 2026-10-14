package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-manager/appv3/manager/request"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReverseProxy struct {
	db  *repository.BaseDB
	log *slog.Logger
}

func NewReverseProxy(db *repository.BaseDB, log *slog.Logger) *ReverseProxy {
	return &ReverseProxy{db: db, log: log}
}

func (rp *ReverseProxy) List(ctx context.Context) ([]*model.ReverseProxy, error) {
	coll := rp.db.ReverseProxy()
	return coll.Find(ctx, bson.M{})
}

func (rp *ReverseProxy) Get(ctx context.Context, loc string) (*model.ReverseProxy, error) {
	coll := rp.db.ReverseProxy()
	return coll.FindOne(ctx, bson.M{"location": loc})
}

func (rp *ReverseProxy) Create(ctx context.Context, param request.ReverseProxyUpsert) error {
	now := time.Now()
	header := param.Header.Canonical()

	doc := &model.ReverseProxy{
		Location:  param.Location,
		Name:      param.Name,
		Manager:   param.Manager,
		Broker:    param.Broker,
		Header:    header,
		Upstream:  param.Upstream,
		CreatedAt: now,
		UpdatedAt: now,
	}

	coll := rp.db.ReverseProxy()
	_, err := coll.InsertOne(ctx, doc)

	return err
}

func (rp *ReverseProxy) Update(ctx context.Context, param request.ReverseProxyUpsert) error {
	now := time.Now()
	header := param.Header.Canonical()

	filter := bson.M{"location": param.Location}
	update := bson.M{"$set": bson.M{
		"name":       param.Name,
		"upstream":   param.Upstream,
		"header":     header,
		"manager":    param.Manager,
		"broker":     param.Broker,
		"updated_at": now,
	}}

	coll := rp.db.ReverseProxy()
	_, err := coll.UpdateOne(ctx, filter, update)

	return err
}

func (rp *ReverseProxy) Delete(ctx context.Context, loc string) error {
	coll := rp.db.ReverseProxy()
	_, err := coll.DeleteOne(ctx, bson.M{"location": loc})

	return err
}
