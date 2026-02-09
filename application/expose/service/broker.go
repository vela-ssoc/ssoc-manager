package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common/store/model"
	"github.com/vela-ssoc/ssoc-common/store/repository"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Broker struct {
	db  repository.Database
	log *slog.Logger
}

func NewBroker(db repository.Database, log *slog.Logger) *Broker {
	return &Broker{db: db, log: log}
}

func (b *Broker) Page(ctx context.Context) (*repository.Pages[model.Broker], error) {
	coll := b.db.Broker()

	return coll.Page(ctx, bson.D{}, 1, 10)
}

func (b *Broker) Create(ctx context.Context, req *request.BrokerCreate) error {
	now := time.Now()
	secret := b.randomSecret(now)

	dat := &model.Broker{
		Name:      req.Name,
		Secret:    secret,
		Exposes:   req.Exposes,
		Config:    req.Config,
		CreatedAt: now,
		UpdatedAt: now,
	}
	coll := b.db.Broker()
	_, err := coll.InsertOne(ctx, dat)

	return err
}

func (*Broker) randomSecret(now time.Time) string {
	buf := make([]byte, 50)
	buf[0] = 0xbe // be -> broker
	binary.BigEndian.PutUint64(buf[1:], uint64(now.UnixNano()))
	_, _ = rand.Read(buf[9:])

	return hex.EncodeToString(buf)
}
