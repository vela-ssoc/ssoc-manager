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

type OAuthConfig struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewOAuthConfig(pdb polydb.Sourcer, log *slog.Logger) *OAuthConfig {
	return &OAuthConfig{pdb: pdb, log: log}
}

func (oat *OAuthConfig) Get(ctx context.Context) (*model.OAuthConfig, error) {
	mdb := oat.pdb.MDB()
	coll := mdb.OAuthConfig()

	return coll.Get(ctx)
}

func (oat *OAuthConfig) Upsert(ctx context.Context, param request.OAuthConfigUpsert) error {
	now := time.Now()
	mod := &model.OAuthConfig{
		Enabled:               param.Enabled,
		ClientID:              param.ClientID,
		ClientSecret:          param.ClientSecret,
		AuthorizationEndpoint: param.AuthorizationEndpoint,
		TokenEndpoint:         param.TokenEndpoint,
		UserinfoEndpoint:      param.UserinfoEndpoint,
		RedirectURIs:          param.RedirectURIs,
		Scopes:                param.Scopes,
		UpdatedAt:             now,
	}

	mdb := oat.pdb.MDB()
	coll := mdb.OAuthConfig()
	opts := options.UpdateOne().SetUpsert(true)
	update := bson.M{"$set": mod, "$setOnInsert": bson.M{"created_at": now}}
	_, err := coll.UpdateOne(ctx, bson.D{}, update, opts)

	return err
}
