package lifestage

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/lifecycle"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewBrokerStatus(pdb polydb.Sourcer, log *slog.Logger) lifecycle.Stage {
	bs := &brokerStatus{pdb: pdb, log: log}
	return lifecycle.Stage{
		Name:    "中继节点状态标记下线",
		OnStart: bs.updateAllOffline,
		OnClose: bs.updateAllOffline,
	}
}

type brokerStatus struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func (bs *brokerStatus) updateAllOffline(ctx context.Context) error {
	mdb := bs.pdb.MDB()
	coll := mdb.Broker()

	filter := bson.M{"status": true}
	update := bson.M{"$set": bson.M{"status": false}}
	_, err := coll.UpdateMany(ctx, filter, update)

	return err
}
