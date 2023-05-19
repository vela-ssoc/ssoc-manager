package push

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
)

type Pusher interface {
	SubstanceTask(ctx context.Context, bids []int64, taskID int64)
}

func NewPush(hub linkhub.Huber) Pusher {
	return &pushImpl{hub: hub}
}

type pushImpl struct {
	hub linkhub.Huber
}

func (pi *pushImpl) SubstanceTask(ctx context.Context, bids []int64, taskID int64) {
	if taskID == 0 || len(bids) == 0 {
		return
	}

	req := &substanceTask{TaskID: taskID}
	futures := pi.hub.Multicast(bids, pathSubstanceTask, req)
	for fut := range futures {
		err := fut.Error()
		if err == nil {
			break
		}
		msg := err.Error()
		tbl := query.SubstanceTask
		_, _ = tbl.WithContext(ctx).
			Where(tbl.TaskID.Eq(taskID)).
			Where(tbl.Executed.Is(false)).
			UpdateColumnSimple(
				tbl.Executed.Value(true),
				tbl.Failed.Value(true),
				tbl.Reason.Value(msg),
			)
	}
}
