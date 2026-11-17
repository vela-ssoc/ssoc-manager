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

type AIChatModel struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewAIChatModel(pdb polydb.Sourcer, log *slog.Logger) *AIChatModel {
	return &AIChatModel{
		pdb: pdb,
		log: log,
	}
}

func (acm *AIChatModel) Get(ctx context.Context) (*model.AIChatModel, error) {
	mdb := acm.pdb.MDB()
	coll := mdb.AIChatModel()

	return coll.Get(ctx)
}

func (acm *AIChatModel) Upsert(ctx context.Context, param request.AIChatModelUpsert) error {
	now := time.Now()
	mdb := acm.pdb.MDB()
	coll := mdb.AIChatModel()

	mod := &model.AIChatModel{
		Enabled:   param.Enabled,
		BaseURL:   param.BaseURL,
		APIKey:    param.APIKey,
		Models:    param.Models,
		Prompt:    param.Prompt,
		UpdatedAt: now,
	}
	update := bson.M{"$set": mod, "$setOnInsert": bson.M{"created_at": now}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, bson.D{}, update, opts)

	return err
}
