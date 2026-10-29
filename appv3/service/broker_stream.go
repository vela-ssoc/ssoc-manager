package service

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/vela-tunnel/muxconn"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BrokerStream struct {
	pdb polydb.Sourcer
	hub muxproto.Huber
	log *slog.Logger
}

func NewBrokerStream(pdb polydb.Sourcer, hub muxproto.Huber, log *slog.Logger) *BrokerStream {
	return &BrokerStream{pdb: pdb, hub: hub, log: log}
}

func (bs *BrokerStream) Substreams(slug string) ([]muxconn.Streamer, bool) {
	peer := bs.hub.Get(slug)
	if peer == nil {
		return nil, false
	}

	mux := peer.MUX()
	streams := mux.Streams()

	return streams, true
}

func (bs *BrokerStream) Close(ctx context.Context, slug string) error {
	if node := bs.hub.Get(slug); node != nil {
		mux := node.MUX()
		_ = mux.Close()
		return nil
	}

	// 连接池中不存在，检查是否数据库状态是否在线，如果在线说明异常退出，
	// 中心端没有来得及标记为下线，状态数据不一致要进行纠错。
	mdb := bs.pdb.MDB()
	coll := mdb.Broker()
	filter := bson.M{"slug": slug, "status": true}
	update := bson.M{"$set": bson.M{"status": false}}
	_, err := coll.UpdateOne(ctx, filter, update)

	return err
}
