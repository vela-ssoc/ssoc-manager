package service

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/memcache"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReverseProxy struct {
	db      polydb.Sourcer
	hub     linkhub.Huber
	log     *slog.Logger
	proxies *memcache.Map[string, *httputil.ReverseProxy]
}

func NewReverseProxy(db polydb.Sourcer, hub linkhub.Huber, log *slog.Logger) *ReverseProxy {
	rp := &ReverseProxy{db: db, hub: hub, log: log}
	rp.proxies = memcache.NewMap(rp.findLocation)

	return rp
}

func (rp *ReverseProxy) List(ctx context.Context) ([]*model.ReverseProxy, error) {
	coll := rp.db.MDB().ReverseProxy()
	return coll.Find(ctx, bson.M{})
}

func (rp *ReverseProxy) Get(ctx context.Context, loc string) (*model.ReverseProxy, error) {
	coll := rp.db.MDB().ReverseProxy()
	return coll.FindOne(ctx, bson.M{"location": loc})
}

func (rp *ReverseProxy) Create(ctx context.Context, param request.ReverseProxyUpsert) error {
	now := time.Now()
	header := param.Header.Canonical()

	loc := param.Location
	doc := &model.ReverseProxy{
		Location:  loc,
		Name:      param.Name,
		Manager:   param.Manager,
		Broker:    param.Broker,
		Header:    header,
		Upstream:  param.Upstream,
		CreatedAt: now,
		UpdatedAt: now,
	}

	coll := rp.db.MDB().ReverseProxy()
	_, err := coll.InsertOne(ctx, doc)
	if err == nil {
		_ = rp.forget(ctx, loc)
	}

	return err
}

func (rp *ReverseProxy) Update(ctx context.Context, param request.ReverseProxyUpsert) error {
	now := time.Now()
	header := param.Header.Canonical()

	loc := param.Location
	filter := bson.M{"location": loc}
	update := bson.M{"$set": bson.M{
		"name":       param.Name,
		"upstream":   param.Upstream,
		"header":     header,
		"manager":    param.Manager,
		"broker":     param.Broker,
		"updated_at": now,
	}}

	coll := rp.db.MDB().ReverseProxy()
	_, err := coll.UpdateOne(ctx, filter, update)
	if err == nil {
		_ = rp.forget(ctx, loc)
	}

	return err
}

func (rp *ReverseProxy) Delete(ctx context.Context, loc string) error {
	coll := rp.db.MDB().ReverseProxy()
	_, err := coll.DeleteOne(ctx, bson.M{"location": loc})
	if err == nil {
		_ = rp.forget(ctx, loc)
	}

	return err
}

func (rp *ReverseProxy) Handle(ctx context.Context, location string, w http.ResponseWriter, r *http.Request) error {
	reverse, err := rp.proxies.Load(ctx, location)
	if err == nil {
		reverse.ServeHTTP(w, r)
	}

	return err
}

func (rp *ReverseProxy) findLocation(ctx context.Context, loc string) (*httputil.ReverseProxy, error) {
	coll := rp.db.MDB().ReverseProxy()
	dat, err := coll.FindOne(ctx, bson.M{"location": loc, "manager": true})
	if err != nil {
		return nil, err
	}
	target, err := url.Parse(dat.Upstream)
	if err != nil {
		return nil, err
	}

	pxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
			pr.SetURL(target)
			for k, v := range dat.Header {
				pr.Out.Header.Set(k, v)
			}
		},
	}

	return pxy, nil
}

func (rp *ReverseProxy) forget(ctx context.Context, loc string) error {
	rp.proxies.Forget(loc)

	_ = rp.hub.Broadcast(ctx, "/api/v3/reverse-proxy", nil)

	return nil
}
