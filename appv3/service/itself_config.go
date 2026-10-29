package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ItselfConfig struct {
	db  polydb.Sourcer
	log *slog.Logger
}

func NewItselfConfig(db polydb.Sourcer, log *slog.Logger) *ItselfConfig {
	return &ItselfConfig{db: db, log: log}
}

func (slf *ItselfConfig) Get(ctx context.Context) (*model.ManagerConfig, error) {
	mdb := slf.db.MDB()
	coll := mdb.ManagerConfig()

	return coll.Get(ctx)
}

func (slf *ItselfConfig) Upsert(ctx context.Context, param request.ItselfConfigUpdate) error {
	now := time.Now()
	mdb := slf.db.MDB()
	coll := mdb.ManagerConfig()

	mod := &model.ManagerConfig{
		Server:    param.Server,
		Logger:    param.Logger,
		UpdatedAt: now,
	}
	update := bson.M{"$set": mod, "$setOnInsert": bson.M{"created_at": now}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, bson.D{}, update, opts)

	return err
}
