package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/integration/ywcmdb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CMDBConfig struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewCMDBConfig(pdb polydb.Sourcer, log *slog.Logger) *CMDBConfig {
	return &CMDBConfig{pdb: pdb, log: log}
}

func (cc *CMDBConfig) Get(ctx context.Context) (*model.CMDBConfig, error) {
	mdb := cc.pdb.MDB()
	coll := mdb.CMDBConfig()

	return coll.Get(ctx)
}

func (cc *CMDBConfig) Upsert(ctx context.Context, param request.CMDBConfigUpsert) error {
	now := time.Now()
	mod := &model.CMDBConfig{
		Endpoint:  param.Endpoint,
		Enabled:   param.Enabled,
		Key:       param.Key,
		Secret:    param.Secret,
		UpdatedAt: now,
	}

	mdb := cc.pdb.MDB()
	coll := mdb.CMDBConfig()
	opts := options.UpdateOne().SetUpsert(true)
	update := bson.M{"$set": mod, "$setOnInsert": bson.M{"created_at": now}}
	_, err := coll.UpdateOne(ctx, bson.D{}, update, opts)

	return err
}

func (cc *CMDBConfig) LoadCMDBConfig(ctx context.Context) (*ywcmdb.Config, error) {
	coll := cc.pdb.MDB().CMDBConfig()
	cfg, err := coll.Get(ctx)
	if err != nil {
		return nil, errcode.ErrCMDBNotConfigured
	} else if !cfg.Enabled {
		return nil, errcode.ErrCMDBDisabled
	}

	return &ywcmdb.Config{
		Endpoint: cfg.Endpoint,
		Key:      cfg.Key,
		Secret:   cfg.Secret,
	}, nil
}
