package service

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/vela-ssoc/vela-manager/param/mrequest"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/integration/elastic"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
)

type ElasticService interface {
	Forward(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	Page(ctx context.Context, page param.Pager) (int64, []*model.Elastic)
	Create(ctx context.Context, ec *mrequest.ElasticCreate) error
	Update(ctx context.Context, eu *mrequest.ElasticUpdate) error
	Delete(ctx context.Context, id int64) error
	Detect(ctx context.Context, host, uname, passwd string) []string
}

func Elastic(qry *query.Query, pusher push.Pusher, forward elastic.Searcher, cfg elastic.Configurer, client netutil.HTTPClient) ElasticService {
	return &elasticService{
		qry:     qry,
		client:  client,
		pusher:  pusher,
		forward: forward,
		cfg:     cfg,
	}
}

type elasticService struct {
	qry     *query.Query
	client  netutil.HTTPClient
	forward elastic.Searcher
	cfg     elastic.Configurer
	pusher  push.Pusher
}

func (biz *elasticService) Forward(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return biz.forward.ServeHTTP(ctx, w, r)
}

func (biz *elasticService) Page(ctx context.Context, page param.Pager) (int64, []*model.Elastic) {
	ret := make([]*model.Elastic, 0, page.Size())
	tbl := biz.qry.Elastic
	db := tbl.WithContext(ctx)
	if kw := page.Keyword(); kw != "" {
		db = db.Where(tbl.Host.Like(kw)).
			Or(tbl.Desc.Like(kw))
	}
	count, err := db.Count()
	if err != nil || count == 0 {
		return 0, ret
	}

	_ = db.Scopes(page.Scope(count)).
		Order(tbl.Enable.Desc()).Order(tbl.ID).
		Scan(&ret)

	for _, m := range ret {
		if len(m.Hosts) == 0 {
			m.Hosts = []string{m.Host}
		}
	}

	return count, ret
}

func (biz *elasticService) Create(ctx context.Context, ec *mrequest.ElasticCreate) error {
	dat := &model.Elastic{
		Host:     ec.Hosts[0],
		Hosts:    ec.Hosts,
		Username: ec.Username,
		Password: ec.Password,
		Desc:     ec.Desc,
		Enable:   ec.Enable,
	}

	enable := ec.Enable
	tbl := biz.qry.Elastic
	if !enable {
		return tbl.WithContext(ctx).Create(dat)
	}

	err := biz.qry.Transaction(func(tx *query.Query) error {
		txdb := tx.WithContext(ctx).Elastic
		if _, err := txdb.Where(tbl.Enable.Is(true)).
			UpdateSimple(tbl.Enable.Value(false)); err != nil {
			return err
		}
		return txdb.Create(dat)
	})
	if err == nil {
		biz.pusher.ElasticReset(ctx)
	}

	return err
}

// Update 更新 es 后端代理
func (biz *elasticService) Update(ctx context.Context, eu *mrequest.ElasticUpdate) error {
	// 先查询原有数据
	id := eu.ID
	tbl := biz.qry.Elastic
	es, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}

	reset := es.Enable || eu.Enable
	type Elastic struct {
		ID       int64    `json:"id,string" gorm:"column:id;primaryKey"` // ID
		Host     string   `json:"host"      gorm:"column:host"`          // es 地址
		Username string   `json:"username"  gorm:"column:username"`      // es 用户名
		Password string   `json:"password"  gorm:"column:password"`      // es 密码
		Hosts    []string `json:"hosts"     gorm:"column:hosts;json"`    // es 服务器
		Desc     string   `json:"desc"      gorm:"column:desc"`          // 简介
		Enable   bool     `json:"enable"    gorm:"column:enable"`        // 是否启用，最多只能有一个启用
	}
	es.Host = eu.Hosts[0]
	es.Username = eu.Username
	es.Password = eu.Password
	es.Hosts = eu.Hosts
	es.Desc = eu.Desc
	es.Enable = eu.Enable
	if !reset {
		return tbl.WithContext(ctx).
			Where(tbl.ID.Eq(id)).
			Save(es)
	}

	err = biz.qry.Transaction(func(tx *query.Query) error {
		db := tx.Elastic.WithContext(ctx)
		if _, exx := db.Where(tbl.Enable.Is(true)).
			Update(tbl.Enable, false); exx != nil {
			return exx
		}
		return db.Where(tbl.ID.Eq(id)).
			Save(es)
	})

	// 是否需要 reset
	if err == nil {
		biz.cfg.Reset()
		biz.pusher.ElasticReset(ctx)
	}

	return err
}

// Delete 根据 ID 删除 es 配置
func (biz *elasticService) Delete(ctx context.Context, id int64) error {
	tbl := biz.qry.Elastic
	db := tbl.WithContext(ctx)
	es, err := db.Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}

	reset := es.Enable
	_, err = db.Delete(es)
	if err == nil && reset {
		biz.cfg.Reset()
		biz.pusher.ElasticReset(ctx)
	}

	return err
}

// Detect 根据 ID 删除 es 配置
//
// https://www.elastic.co/guide/en/elasticsearch/reference/8.3/cat-nodes.html
func (biz *elasticService) Detect(parent context.Context, addr, uname, passwd string) []string {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	addr = strings.TrimRight(addr, "/")
	addr += "/_cat/nodes?format=json&h=http"
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(uname+":"+passwd))
	headers := http.Header{"Authorization": []string{auth}}

	var peers mrequest.ElasticDetects
	_ = biz.client.JSON(ctx, http.MethodGet, addr, nil, &peers, headers)
	ret := peers.Addrs()

	return ret
}
