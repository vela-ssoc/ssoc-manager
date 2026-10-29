package muxroute

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/vela-tunnel/muxconn"
)

type AgentRouter struct {
	db  polydb.Sourcer
	hub muxproto.Huber
	log *slog.Logger
}

func NewAgentRouter(db polydb.Sourcer, hub muxproto.Huber, log *slog.Logger) *AgentRouter {
	return &AgentRouter{
		db:  db,
		hub: hub,
		log: log,
	}
}

func (r *AgentRouter) TLD() string {
	return muxproto.AgentTLD
}

func (r *AgentRouter) MUX(ctx context.Context, host string) (muxconn.Muxer, error) {
	args := []any{"host", host}
	id, err := strconv.ParseInt(host, 10, 64)
	if err != nil {
		r.log.Info("输入的节点标识无效", args...)
		return nil, errcode.ErrInvalidAgentNodeIdentify
	}

	qry := r.db.RDB()
	dao := qry.Minion
	node, err := dao.WithContext(ctx).
		Select(dao.ID, dao.BrokerSlug, dao.BrokerName, dao.Status).
		Where(dao.ID.Eq(id)).
		First()
	if err != nil {
		args = append(args, "err", err)
		r.log.Info("数据库查询节点出错", args...)
		return nil, errcode.ErrNodeNotExist
	}
	if node.Status != model.MSOnline || node.BrokerSlug == "" {
		args = append(args, "status", node.Status)
		r.log.Info("节点已离线", args...)
		return nil, errcode.ErrAgentNodeOffline
	}

	brok := r.hub.Get(node.BrokerSlug)
	if brok == nil {
		args = append(args, "slug", node.BrokerSlug)
		r.log.Info("所在中继节点已离线", args...)
		return nil, errcode.ErrBrokerNodeOffline
	}

	return brok.MUX(), nil
}
