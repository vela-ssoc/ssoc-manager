package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Broker struct {
	db  polydb.Sourcer
	log *slog.Logger
}

func NewBroker(db polydb.Sourcer, log *slog.Logger) *Broker {
	return &Broker{db: db, log: log}
}

func (bk *Broker) Get(ctx context.Context, slug string) (*model.Broker, error) {
	coll := bk.db.MDB().Broker()
	return coll.FindOne(ctx, bson.M{"slug": slug})
}

func (bk *Broker) List(ctx context.Context) ([]*model.Broker, error) {
	coll := bk.db.MDB().Broker()
	return coll.Find(ctx, bson.M{})
}

func (bk *Broker) Create(ctx context.Context, param request.BrokerUpsert) error {
	now := time.Now()
	tmp := make([]byte, 50)
	_, _ = rand.Read(tmp)
	secret := "broker." + hex.EncodeToString(tmp)

	doc := &model.Broker{
		Slug:      param.Slug,
		Name:      param.Name,
		Secret:    secret,
		Exposes:   param.Exposes,
		Config:    param.Config,
		CreatedAt: now,
		UpdatedAt: now,
	}

	coll := bk.db.MDB().Broker()
	_, err := coll.InsertOne(ctx, doc)

	return err
}

func (bk *Broker) Update(ctx context.Context, param request.BrokerUpsert) error {
	now := time.Now()
	filter := bson.M{"slug": param.Slug}
	update := bson.M{"$set": bson.M{
		"name":       param.Name,
		"exposes":    param.Exposes,
		"config":     param.Config,
		"updated_at": now,
	}}

	coll := bk.db.MDB().Broker()
	_, err := coll.UpdateOne(ctx, filter, update)

	return err
}

func (bk *Broker) Delete(ctx context.Context, slug string) error {
	filter := bson.M{"slug": slug}

	coll := bk.db.MDB().Broker()
	_, err := coll.DeleteOne(ctx, filter)

	return err
}

func (bk *Broker) Ping(ctx context.Context, peer muxproto.Node) error {
	now := time.Now()
	slug := peer.Host()
	filter := bson.M{"slug": slug, "status": true}
	update := bson.M{"$set": bson.M{"heartbeat_at": now}}

	coll := bk.db.MDB().Broker()
	_, err := coll.UpdateOne(ctx, filter, update)

	return err
}
