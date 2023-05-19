package service

import (
	"context"
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/integrate"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type ElasticService interface {
	Forward(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	Page(ctx context.Context, page param.Pager) (int64, []*model.Elastic)
	Create(ctx context.Context, ec *param.ElasticCreate) error
	Update(ctx context.Context, eu *param.ElasticUpdate) error
	Delete(ctx context.Context, id int64) error
}

func Elastic(name string) ElasticService {
	px := integrate.NewElastic(name)
	return &elasticService{
		px: px,
	}
}

type elasticService struct {
	px integrate.ElasticSearcher
}

func (ela *elasticService) Forward(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := ela.px.Load(ctx)
	if err == nil {
		h.ServeHTTP(w, r)
	}
	return err
}

func (ela *elasticService) Page(ctx context.Context, page param.Pager) (int64, []*model.Elastic) {
	ret := make([]*model.Elastic, 0, page.Size())
	tbl := query.Elastic
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

	return count, ret
}

func (ela *elasticService) Create(ctx context.Context, ec *param.ElasticCreate) error {
	dat := &model.Elastic{
		Host:     ec.Host,
		Username: ec.Username,
		Password: ec.Password,
		Desc:     ec.Desc,
		Enable:   ec.Enable,
	}

	tbl := query.Elastic
	if !ec.Enable {
		return tbl.WithContext(ctx).Create(dat)
	}

	db := tbl.WithContext(ctx).UnderlyingDB()
	err := query.Use(db).Transaction(func(tx *query.Query) error {
		txe := tx.WithContext(ctx).Elastic
		if _, err := txe.Where(tbl.Enable.Is(true)).
			Update(tbl.Enable, false); err != nil {
			return err
		}
		return txe.Create(dat)
	})
	if err == nil {
		ela.px.Reset()
	}

	return err
}

// Update 更新 es 后端代理
func (ela *elasticService) Update(ctx context.Context, eu *param.ElasticUpdate) error {
	// 先查询原有数据
	tbl := query.Elastic
	es, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(eu.ID)).First()
	if err != nil {
		return err
	}

	// 更新数据
	reset := es.Enable || eu.Enable
	es.Host = eu.Host
	es.Username = eu.Username
	es.Password = eu.Password
	es.Enable = eu.Enable
	es.Desc = eu.Desc

	if !eu.Enable {
		_, err = tbl.WithContext(ctx).Updates(es)
	} else {
		err = query.Q.Transaction(func(tx *query.Query) error {
			db := tx.Elastic.WithContext(ctx)
			if _, exx := db.Where(tbl.Enable.Is(true)).
				Update(tbl.Enable, false); exx != nil {
				return exx
			}

			_, exx := db.Updates(es)
			return exx
		})
	}

	// 是否需要 reset
	if err == nil && reset {
		ela.px.Reset()
	}

	return err
}

// Delete 根据 ID 删除 es 配置
func (ela *elasticService) Delete(ctx context.Context, id int64) error {
	tbl := query.Elastic
	db := tbl.WithContext(ctx)
	es, err := db.Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}

	reset := es.Enable
	_, err = db.Delete(es)
	if err == nil && reset {
		ela.px.Reset()
	}

	return err
}
