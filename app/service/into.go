package service

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/problem"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type IntoService interface {
	BRR(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error
	BWS(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error
	ARR(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error
	AWS(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error
}

func Into(hub linkhub.Huber) IntoService {
	name := hub.Name()
	ito := &intoService{
		name: name,
		hub:  hub,
	}
	upgrade := netutil.Upgrade(ito.upgradeErrorFunc)
	ito.upgrade = upgrade

	return ito
}

type intoService struct {
	name    string
	hub     linkhub.Huber
	upgrade websocket.Upgrader
}

func (ito *intoService) BRR(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error {
	bid, err := strconv.ParseInt(node, 10, 64)
	if err != nil {
		return errcode.ErrNodeNotExist
	}

	brk := query.Broker
	broker, err := brk.WithContext(ctx).
		Select(brk.ID).
		Where(brk.Status.Is(true)).
		Where(brk.ID.Eq(bid)).
		First()
	if err != nil {
		return errcode.ErrNodeNotExist
	}

	ito.hub.Forward(broker.ID, w, r)

	return nil
}

func (ito *intoService) BWS(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error {
	bid, err := strconv.ParseInt(node, 10, 64)
	if err != nil {
		return errcode.ErrNodeNotExist
	}

	brk := query.Broker
	broker, err := brk.WithContext(ctx).
		Select(brk.ID).
		Where(brk.Status.Is(true)).
		Where(brk.ID.Eq(bid)).
		First()
	if err != nil {
		return errcode.ErrNodeNotExist
	}

	path := r.URL.Path
	up, _, err := ito.hub.Stream(ctx, broker.ID, path, nil)
	if err != nil {
		return err
	}

	down, err := ito.upgrade.Upgrade(w, r, nil)
	if err != nil {
		_ = up.Close()
		return err
	}

	netutil.PipeWebsocket(up, down)

	return nil
}

func (ito *intoService) ARR(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error {
	mon := query.Minion
	db := mon.WithContext(ctx).
		Select(mon.ID, mon.BrokerID).
		Where(mon.Inet.Eq(node))
	if mid, _ := strconv.ParseInt(node, 10, 64); mid != 0 {
		db.Or(mon.ID.Eq(mid))
	}
	minion, err := db.First()
	if err != nil {
		return errcode.ErrNodeNotExist
	}

	r.Header.Set(linkhub.HeaderXNodeID, strconv.FormatInt(minion.ID, 10))
	ito.hub.Forward(minion.BrokerID, w, r)

	return nil
}

func (ito *intoService) AWS(ctx context.Context, w http.ResponseWriter, r *http.Request, node string) error {
	mon := query.Minion
	db := mon.WithContext(ctx).
		Select(mon.ID, mon.BrokerID).
		Where(mon.Inet.Eq(node))
	if mid, _ := strconv.ParseInt(node, 10, 64); mid != 0 {
		db.Or(mon.ID.Eq(mid))
	}
	minion, err := db.First()
	if err != nil {
		return errcode.ErrNodeNotExist
	}

	header := http.Header{linkhub.HeaderXNodeID: []string{strconv.FormatInt(minion.ID, 10)}}
	path := r.URL.Path
	up, _, err := ito.hub.Stream(ctx, minion.BrokerID, path, header)
	if err != nil {
		return err
	}

	down, err := ito.upgrade.Upgrade(w, r, nil)
	if err != nil {
		_ = up.Close()
		return err
	}

	netutil.PipeWebsocket(up, down)

	return nil
}

func (ito *intoService) upgradeErrorFunc(w http.ResponseWriter, r *http.Request, status int, reason error) {
	pd := &problem.Detail{
		Type:     ito.name,
		Title:    "websocket 协议升级错误",
		Status:   status,
		Detail:   reason.Error(),
		Instance: r.RequestURI,
	}
	_ = pd.JSON(w)
}
