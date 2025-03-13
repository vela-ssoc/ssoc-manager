package taskexec

import (
	"context"
	"database/sql"
	"iter"

	"github.com/vela-ssoc/ssoc-manager/app/service/internal/minionfilter"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"gorm.io/gen"
)

func newQueryMode(qry *query.Query, flt *minionfilter.Filter, f model.TaskExecuteFilter, excludes []string) matcher {
	var inputs request.CondWhereInputs
	for _, filter := range f.Filters {
		inputs.Filters = append(inputs.Filters, &request.CondWhereInput{
			Key:      filter.Key,
			Operator: filter.Operator,
			Value:    filter.Value,
		})
	}

	args := &request.KeywordConditions{
		Keywords: request.Keywords{Keyword: f.Keyword},
		Conditions: request.Conditions{
			CondWhereInputs: inputs,
		},
	}
	wheres, err := flt.Wheres(args, excludes)

	tbl := qry.Minion
	wheres = append(wheres, tbl.Status.Neq(uint8(model.MSDelete)))

	return &queryModeMatch{
		qry:    qry,
		wheres: wheres,
		err:    err,
	}
}

type queryModeMatch struct {
	qry    *query.Query
	wheres []gen.Condition
	err    error
}

func (qmm *queryModeMatch) Count(ctx context.Context) (int64, error) {
	if err := qmm.err; err != nil {
		return 0, err
	}

	minion, minionTag := qmm.qry.Minion, qmm.qry.MinionTag
	minionDo, minionTagDo := minion.WithContext(ctx), minionTag.WithContext(ctx)
	return minionDo.Distinct(minion.ID).
		LeftJoin(minionTagDo, minion.ID.EqCol(minionTag.MinionID)).
		Where(qmm.wheres...).
		Count()
}

func (qmm *queryModeMatch) Iter(ctx context.Context, batchSize int) iter.Seq2[[]*model.Minion, error] {
	minion, minionTag := qmm.qry.Minion, qmm.qry.MinionTag
	minionTagDo := minionTag.WithContext(ctx)
	minionDo := minion.WithContext(ctx).
		Distinct(minion.ID).
		LeftJoin(minionTagDo, minion.ID.EqCol(minionTag.MinionID)).
		Where(qmm.wheres...)
	return func(yield func([]*model.Minion, error) bool) {
		var buf []*model.Minion
		_ = minionDo.FindInBatches(&buf, batchSize, func(tx gen.Dao, _ int) error {
			var minionIDs []int64
			for _, m := range buf {
				minionIDs = append(minionIDs, m.ID)
			}

			minions, err := minion.WithContext(ctx).
				Where(minion.ID.In(minionIDs...)).
				Find()

			if !yield(minions, err) {
				return sql.ErrNoRows
			}

			return nil
		})
	}
}
