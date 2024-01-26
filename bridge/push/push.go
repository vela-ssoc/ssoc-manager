package push

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/xgfone/ship/v5"

	"github.com/vela-ssoc/vela-common-mb/accord"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"gorm.io/gen/field"
)

type Pusher interface {
	TaskTable(ctx context.Context, bids []int64, tid int64)
	TaskSync(ctx context.Context, bid int64, mids []int64)
	TaskDiff(ctx context.Context, bid, mid, sid int64, inet string)
	ThirdUpdate(ctx context.Context, name string)
	ThirdDelete(ctx context.Context, name string)
	ElasticReset(ctx context.Context)
	EmcReset(ctx context.Context)
	EmailReset(ctx context.Context)
	StoreReset(ctx context.Context, id string)
	NotifierReset(ctx context.Context)
	Startup(ctx context.Context, bid, mid int64)
	Upgrade(ctx context.Context, bid int64, mid []int64, semver, customized string)
	Command(ctx context.Context, bid int64, mids []int64, cmd string)
	SavePprof(ctx context.Context, bid, mid int64, second int, dest string) error
}

func NewPush(hub linkhub.Huber) Pusher {
	return &pushImpl{hub: hub}
}

type pushImpl struct {
	hub linkhub.Huber
}

func (pi *pushImpl) TaskTable(_ context.Context, bids []int64, tid int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	req := &accord.TaskTable{TaskID: tid}
	ret := pi.hub.Multicast(nil, bids, accord.FPTaskTable, req)
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

func (pi *pushImpl) TaskSync(_ context.Context, bid int64, mids []int64) {
	if bid == 0 || len(mids) == 0 {
		return
	}
	req := &accord.IDs{ID: mids}
	_ = pi.hub.Oneway(nil, bid, accord.FPTaskSync, req)
}

func (pi *pushImpl) TaskDiff(ctx context.Context, bid, mid, sid int64, inet string) {
	if bid == 0 || mid == 0 || sid == 0 || inet == "" {
		return
	}
	req := &accord.TaskLoadRequest{MinionID: mid, SubstanceID: sid, Inet: inet}
	_ = pi.hub.Oneway(nil, bid, accord.FPTaskLoad, req)
}

func (pi *pushImpl) ThirdUpdate(ctx context.Context, name string) {
	pi.thirdDiff(ctx, name, accord.ThirdUpdate)
}

func (pi *pushImpl) ThirdDelete(ctx context.Context, name string) {
	pi.thirdDiff(ctx, name, accord.ThirdDelete)
}

func (pi *pushImpl) ElasticReset(ctx context.Context) {
	pi.hub.Broadcast(nil, accord.FPElasticReset, nil)
}

func (pi *pushImpl) EmcReset(ctx context.Context) {
	pi.hub.Broadcast(nil, accord.FPEmcReset, nil)
}

func (pi *pushImpl) EmailReset(ctx context.Context) {
	pi.hub.Broadcast(nil, accord.FPEmailReset, nil)
}

func (pi *pushImpl) StoreReset(ctx context.Context, id string) {
	req := &accord.StoreRestRequest{ID: id}
	pi.hub.Broadcast(nil, accord.FPStoreReset, req)
}

func (pi *pushImpl) NotifierReset(ctx context.Context) {
	pi.hub.Broadcast(nil, accord.FPNotifierReset, nil)
}

func (pi *pushImpl) Startup(ctx context.Context, bid int64, mid int64) {
	req := accord.Startup{ID: mid}
	_ = pi.hub.Oneway(nil, bid, accord.FPStartup, req)
}

func (pi *pushImpl) Upgrade(ctx context.Context, bid int64, mids []int64, semver, customized string) {
	req := accord.Upgrade{ID: mids, Semver: semver, Customized: customized}
	_ = pi.hub.Oneway(nil, bid, accord.FPUpgrade, req)
}

func (pi *pushImpl) Command(ctx context.Context, bid int64, mids []int64, cmd string) {
	req := accord.Command{ID: mids, Cmd: cmd}
	_ = pi.hub.Oneway(nil, bid, accord.FPCommand, req)
}

func (pi *pushImpl) thirdDiff(ctx context.Context, name, event string) {
	req := &accord.ThirdDiff{Name: name, Event: event}
	pi.hub.Broadcast(nil, accord.FPThirdDiff, req)
}

func (pi *pushImpl) SavePprof(ctx context.Context, bid, mid int64, second int, dest string) error {
	if second <= 0 {
		second = 30
	}

	strURL := fmt.Sprintf("/api/v1/arr/pprof/profile?seconds=%d", second)
	header := http.Header{linkhub.HeaderXNodeID: []string{strconv.FormatInt(mid, 10)}}

	res, err := pi.hub.Do(ctx, bid, http.MethodGet, strURL, nil, header)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	code := res.StatusCode
	if code/100 != 2 {
		return ship.ErrBadRequest
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()

	_, err = io.Copy(f, res.Body)

	return err
}
