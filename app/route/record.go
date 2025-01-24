package route

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
)

type Recorder interface {
	Save(context.Context, *model.Oplog) error
}

func NewRecord(qry *query.Query) Recorder {
	return &record{
		qry: qry,
	}
}

type record struct {
	qry *query.Query
}

func (rcd *record) Save(ctx context.Context, oplog *model.Oplog) error {
	return rcd.qry.Oplog.WithContext(ctx).Create(oplog)
}
