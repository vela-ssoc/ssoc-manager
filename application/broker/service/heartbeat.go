package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common/store/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Heartbeat struct {
	db  repository.Database
	log *slog.Logger
}

func NewHeartbeat(db repository.Database, log *slog.Logger) *Heartbeat {
	return &Heartbeat{
		db:  db,
		log: log,
	}
}

// Ping 处理 broker 节点发来的心跳包。
func (hb *Heartbeat) Ping(ctx context.Context, id bson.ObjectID) error {
	now := time.Now()
	update := bson.M{"$set": bson.M{
		"tunnel_stat.keepalive_at": now,
	}}

	coll := hb.db.Broker()
	_, err := coll.UpdateByID(ctx, id, update)

	return err
}
