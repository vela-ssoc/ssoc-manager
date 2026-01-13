package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/application/expose/response"
)

type ZombieConnect struct {
	qry *query.Query
	log *slog.Logger
}

func NewZombieConnect(qry *query.Query, log *slog.Logger) *ZombieConnect {
	return &ZombieConnect{
		qry: qry,
		log: log,
	}
}

// Page 分页查询僵尸连接。
//
// 通过 agent 最后一次心跳时间和当前时间的距离来判断是否是僵尸节点。
func (zc *ZombieConnect) Page(ctx context.Context, req *request.Pages) (*response.Pages[model.Minion], error) {
	tbl := zc.qry.Minion
	dao := tbl.WithContext(ctx)

	now := time.Now()
	heartAt := now.Add(-10 * time.Minute) // 这是一个经验值
	online := uint8(model.MSOnline)
	dao = dao.Where(tbl.Status.Eq(online), tbl.HeartbeatAt.Lte(heartAt))

	res := response.NewPages[model.Minion](req.PageSize())
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return res, nil
	}

	dats, err := dao.Scopes(res.Scope(cnt)).Find()
	if err != nil {
		return nil, err
	}

	return res.SetRecords(dats), nil
}
