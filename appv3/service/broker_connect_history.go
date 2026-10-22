package service

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BrokerConnectHistory struct {
	db  polydb.Sourcer
	log *slog.Logger
}

func NewBrokerConnectHistory(db polydb.Sourcer, log *slog.Logger) *BrokerConnectHistory {
	return &BrokerConnectHistory{db: db, log: log}
}

func (bch *BrokerConnectHistory) Page(ctx context.Context, slug string, param request.PageSize) (*repository.Pages[model.BrokerConnectHistory], error) {
	filter := bson.M{"slug": slug}
	order := bson.D{{Key: "disconnected_at", Value: -1}}
	opts := options.Find().SetSort(order)

	coll := bch.db.MDB().BrokerConnectHistory()
	pn, sn := param.PSN()

	return coll.Page(ctx, filter, pn, sn, opts)
}
