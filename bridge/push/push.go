package push

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/accord"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"gorm.io/gen/field"
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
	NotifierReset(ctx context.Context)
	CmdbReset(ctx context.Context)
	Startup(ctx context.Context, bid, mid int64)
}

func NewPush(hub linkhub.Huber) Pusher {
	return &pushImpl{hub: hub}
}

type pushImpl struct {
	hub linkhub.Huber
}

func (pi *pushImpl) TaskTable(ctx context.Context, bids []int64, tid int64) {
	req := &accord.TaskTable{TaskID: tid}
	ret := pi.hub.Multicast(bids, accord.FPTaskTable, req)
	tbl := query.SubstanceTask
	for ft := range ret {
		err := ft.Error()
		if err == nil {
			continue
		}

		assigns := []field.AssignExpr{
			tbl.Executed.Value(true),
			tbl.Reason.Value(err.Error()),
			tbl.Failed.Value(true),
		}
		bid := ft.BrokerID()
		_, _ = tbl.WithContext(ctx).
			Where(tbl.TaskID.Eq(tid), tbl.BrokerID.Eq(bid)).
			UpdateColumnSimple(assigns...)
	}
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

func (pi *pushImpl) NotifierReset(ctx context.Context) {
	pi.hub.Broadcast(accord.FPNotifierReset, nil)
}

func (pi *pushImpl) CmdbReset(ctx context.Context) {
	pi.hub.Broadcast(accord.FPCmdbReset, nil)
}

func (pi *pushImpl) Startup(ctx context.Context, bid int64, mid int64) {
	req := accord.Startup{ID: mid}
	_ = pi.hub.Oneway(bid, accord.FPStartup, req)
}

func (pi *pushImpl) thirdDiff(ctx context.Context, name, event string) {
	req := &accord.ThirdDiff{Name: name, Event: event}
	pi.hub.Broadcast(accord.FPThirdDiff, req)
}
