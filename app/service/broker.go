package service

import (
	"context"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type BrokerService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.Broker)
	Indices(ctx context.Context, idx param.Indexer) []*model.Broker
	Create(ctx context.Context, req *param.BrokerCreate) error
	Update(ctx context.Context, req *param.BrokerUpdate) error
	Delete(ctx context.Context, id int64) error
}

func Broker() BrokerService {
	nano := time.Now().UnixNano()
	random := rand.New(rand.NewSource(nano))
	return &brokerService{
		random: random,
	}
}

type brokerService struct {
	random *rand.Rand
}

func (biz *brokerService) Page(ctx context.Context, page param.Pager) (int64, []*model.Broker) {
	tbl := query.Broker
	dao := tbl.WithContext(ctx)
	if kw := page.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw)).
			Or(tbl.Servername.Like(kw))
	}
	count, err := dao.Count()
	if count == 0 || err != nil {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *brokerService) Indices(ctx context.Context, idx param.Indexer) []*model.Broker {
	tbl := query.Broker
	dao := tbl.WithContext(ctx)
	if kw := idx.Keyword(); kw != "" {
		dao.Or(tbl.Name.Like(kw), tbl.Servername.Like(kw))
	}

	dats, _ := dao.Scopes(idx.Scope).Find()

	return dats
}

func (biz *brokerService) Create(ctx context.Context, req *param.BrokerCreate) error {
	buf := make([]byte, 20)
	biz.random.Read(buf)
	secret := hex.EncodeToString(buf)

	now := time.Now()
	brk := &model.Broker{
		ID:          0,
		Name:        req.Name,
		Servername:  req.Servername,
		LAN:         req.LAN,
		VIP:         req.VIP,
		Secret:      secret,
		Bind:        req.Bind,
		HeartbeatAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return query.Broker.
		WithContext(ctx).
		Create(brk)
}

func (biz *brokerService) Update(ctx context.Context, req *param.BrokerUpdate) error {
	tbl := query.Broker
	brk, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		First()
	if err != nil {
		return err
	}
	brk.Name = req.Name
	brk.Bind = req.Bind
	brk.Servername = req.Servername
	brk.LAN = req.LAN
	brk.VIP = req.VIP

	return tbl.WithContext(ctx).
		Save(brk)
}

func (biz *brokerService) Delete(ctx context.Context, id int64) error {
	// 查询节点是否在线，在线的节点目前不允许删除
	tbl := query.Broker
	brk, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if brk.Status {
		return errcode.ErrNodeStatus
	}

	_, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()

	return err
}
