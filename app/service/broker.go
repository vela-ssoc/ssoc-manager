package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type BrokerService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.Broker)
	Indices(ctx context.Context, idx param.Indexer) []*model.Broker
}

func Broker() BrokerService {
	return &brokerService{}
}

type brokerService struct{}

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
