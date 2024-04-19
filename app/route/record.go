package route

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
)

type Recorder interface {
	Save(context.Context, *model.Oplog) error
}

func NewRecord() Recorder {
	return &record{}
}

type record struct{}

func (*record) Save(ctx context.Context, oplog *model.Oplog) error {
	return query.Oplog.WithContext(ctx).Create(oplog)
}
