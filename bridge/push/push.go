package push

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/accord"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
)

type Pusher interface {
	TaskTable(ctx context.Context, bids []int64, tid int64)
	TaskSync(ctx context.Context, bid, mid int64, inet string)
	TaskDiff(ctx context.Context, bid, mid, sid int64, inet string)
	ThirdUpdate(ctx context.Context, name string)
	ThirdDelete(ctx context.Context, name string)
	ElasticReset(ctx context.Context)
	EmcReset(ctx context.Context)
	StoreReset(ctx context.Context, id string)
}

func NewPush(hub linkhub.Huber) Pusher {
	return &pushImpl{hub: hub}
}

type pushImpl struct {
	hub linkhub.Huber
}

func (pi *pushImpl) TaskTable(ctx context.Context, bids []int64, tid int64) {
}

func (pi *pushImpl) TaskSync(ctx context.Context, bid, mid int64, inet string) {
	if bid == 0 || mid == 0 || inet == "" {
		return
	}
	req := &accord.TaskSyncRequest{MinionID: mid, Inet: inet}
	_ = pi.hub.Oneway(bid, accord.FPTaskSync, req)
}

func (pi *pushImpl) TaskDiff(ctx context.Context, bid, mid, sid int64, inet string) {
	if bid == 0 || mid == 0 || sid == 0 || inet == "" {
		return
	}
	req := &accord.TaskLoadRequest{MinionID: mid, SubstanceID: sid, Inet: inet}
	_ = pi.hub.Oneway(bid, accord.FPTaskLoad, req)
}

func (pi *pushImpl) ThirdUpdate(ctx context.Context, name string) {
	pi.thirdDiff(ctx, name, accord.ThirdUpdate)
}

func (pi *pushImpl) ThirdDelete(ctx context.Context, name string) {
	pi.thirdDiff(ctx, name, accord.ThirdDelete)
}

func (pi *pushImpl) ElasticReset(ctx context.Context) {
	pi.hub.Broadcast(accord.FPElasticReset, nil)
}

func (pi *pushImpl) EmcReset(ctx context.Context) {
	pi.hub.Broadcast(accord.FPEmcReset, nil)
}

func (pi *pushImpl) StoreReset(ctx context.Context, id string) {
	req := &accord.StoreRestRequest{ID: id}
	pi.hub.Broadcast(accord.FPStoreReset, req)
}

func (pi *pushImpl) thirdDiff(ctx context.Context, name, event string) {
	req := &accord.ThirdDiff{Name: name, Event: event}
	pi.hub.Broadcast(accord.FPThirdDiff, req)
}
