package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common/store/repository"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Broker struct {
	db  repository.Database
	cfg config.Database
	log *slog.Logger
}

func NewBroker(db repository.Database, cfg config.Database, log *slog.Logger) *Broker {
	return &Broker{
		db:  db,
		cfg: cfg,
		log: log,
	}
}

func (bs *Broker) Reset(timeout time.Duration) error {
	filter := bson.M{"status": true}
	update := bson.M{"$set": bson.M{"status": false}}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	coll := bs.db.Broker()
	_, err := coll.UpdateMany(ctx, filter, update)

	return err
}

func (bs *Broker) LoadBoot(context.Context) (*muxproto.BrokerBootConfig, error) {
	return &muxproto.BrokerBootConfig{URI: bs.cfg.URI}, nil
}
