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
	Page(ctx context.Context, page param.Pager) (int64, param.BrokerSummaries)
	Indices(ctx context.Context, idx param.Indexer) []*model.Broker
	Create(ctx context.Context, req *param.BrokerCreate) error
	Update(ctx context.Context, req *param.BrokerUpdate) error
	Delete(ctx context.Context, id int64) error
	Goos(ctx context.Context) []*param.BrokerGoos
	Stats(ctx context.Context) ([]*model.BrokerStat, error)
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

func (biz *brokerService) Page(ctx context.Context, page param.Pager) (int64, param.BrokerSummaries) {
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

	var ret param.BrokerSummaries
	_ = dao.Scopes(page.Scope(count)).Scan(&ret)
	certIDs, certMap := ret.CertMap()
	if len(certIDs) == 0 || len(certMap) == 0 {
		return count, ret
	}

	certTbl := query.Certificate
	certs, _ := certTbl.WithContext(ctx).
		Omit(certTbl.Certificate, certTbl.PrivateKey).
		Where(certTbl.ID.In(certIDs...)).
		Find()
	for _, cert := range certs {
		certID := cert.ID
		summaries := certMap[certID]
		for _, sm := range summaries {
			sm.Certificate = cert
		}
	}

	return count, ret
}

func (biz *brokerService) Indices(ctx context.Context, idx param.Indexer) []*model.Broker {
	tbl := query.Broker
	dao := tbl.WithContext(ctx).Order(tbl.ID)
	if kw := idx.Keyword(); kw != "" {
		dao.Or(tbl.Name.Like(kw), tbl.Servername.Like(kw))
	}

	dats, _ := dao.Scopes(idx.Scope).Find()

	return dats
}

func (biz *brokerService) Create(ctx context.Context, req *param.BrokerCreate) error {
	if certID := req.CertID; certID != 0 {
		tbl := query.Certificate
		count, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(certID)).Count()
		if err != nil || count == 0 {
			return errcode.ErrCertificate
		}
	}

	buf := make([]byte, 20)
	biz.random.Read(buf)
	secret := hex.EncodeToString(buf)

	now := time.Now()
	brk := &model.Broker{
		Name:        req.Name,
		Servername:  req.Servername,
		LAN:         req.LAN,
		VIP:         req.VIP,
		Secret:      secret,
		CertID:      req.CertID,
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
	if certID := req.CertID; certID != 0 {
		tbl := query.Certificate
		count, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(certID)).Count()
		if err != nil || count == 0 {
			return errcode.ErrCertificate
		}
	}

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
	brk.CertID = req.CertID

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

func (biz *brokerService) Goos(ctx context.Context) []*param.BrokerGoos {
	strSQL := "SELECT broker_id AS id, " +
		"COUNT(IF(goos = 'linux', TRUE, NULL))   AS linux,   " +
		"COUNT(IF(goos = 'windows', TRUE, NULL)) AS windows, " +
		"COUNT(IF(goos = 'darwin', TRUE, NULL))  AS darwin   " +
		" FROM minion" +
		" GROUP BY broker_id "

	ret := make([]*param.BrokerGoos, 0, 10)
	query.Minion.WithContext(ctx).UnderlyingDB().Raw(strSQL).Scan(&ret)

	size := len(ret)
	if size == 0 {
		return ret
	}

	index := make(map[int64]*param.BrokerGoos, size)
	bids := make([]int64, 0, size)
	for _, gc := range ret {
		bid := gc.ID
		if bid == 0 {
			continue
		}
		if _, ok := index[bid]; ok {
			continue
		}

		index[bid] = gc
		bids = append(bids, bid)
	}

	if len(bids) != 0 {
		tbl := query.Broker
		brks, _ := tbl.WithContext(ctx).
			Select(tbl.ID, tbl.Name).
			Where(tbl.ID.In(bids...)).
			Find()
		for _, brk := range brks {
			id, name := brk.ID, brk.Name
			if gc := index[id]; gc != nil {
				gc.Name = name
			}
		}
	}

	return ret
}

func (biz *brokerService) Stats(ctx context.Context) ([]*model.BrokerStat, error) {
	tbl := query.BrokerStat
	return tbl.WithContext(ctx).Order(tbl.ID).Limit(100).Find()
}
