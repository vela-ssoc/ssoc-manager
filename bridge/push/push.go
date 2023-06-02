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

func (pi *pushImpl) thirdDiff(ctx context.Context, name, event string) {
	req := &accord.ThirdDiff{Name: name, Event: event}
	pi.hub.Broadcast(accord.FPThirdDiff, req)
}

//
//func (pi *pushImpl) SubstanceTask(ctx context.Context, bids []int64, taskID int64) {
//	if taskID == 0 || len(bids) == 0 {
//		return
//	}
//
//	req := &substanceTask{TaskID: taskID}
//	futures := pi.hub.Multicast(bids, pathSubstanceTask, req)
//	for fut := range futures {
//		err := fut.Error()
//		if err == nil {
//			break
//		}
//		msg := err.Error()
//		tbl := query.SubstanceTask
//		_, _ = tbl.WithContext(ctx).
//			Where(tbl.TaskID.Eq(taskID)).
//			Where(tbl.Executed.Is(false)).
//			UpdateColumnSimple(
//				tbl.Executed.Value(true),
//				tbl.Failed.Value(true),
//				tbl.Reason.Value(msg),
//			)
//	}
//}
