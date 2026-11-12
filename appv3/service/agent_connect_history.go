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

type AgentConnectHistory struct {
	db  polydb.Sourcer
	log *slog.Logger
}

func NewAgentConnectHistory(db polydb.Sourcer, log *slog.Logger) *AgentConnectHistory {
	return &AgentConnectHistory{db: db, log: log}
}

func (ach *AgentConnectHistory) Page(ctx context.Context, id int64, param request.PageSize) (*repository.Pages[model.AgentConnectHistory], error) {
	filter := bson.M{"agent_idv1": id}
	order := bson.D{{Key: "disconnected_at", Value: -1}}
	opts := options.Find().SetSort(order)

	coll := ach.db.MDB().AgentConnectHistory()
	pn, sn := param.PSN()

	return coll.Page(ctx, filter, pn, sn, opts)
}
