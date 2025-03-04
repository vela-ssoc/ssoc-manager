package taskexec

import (
	"context"
	"database/sql"
	"iter"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"gorm.io/gen"
)

func newInetMode(qry *query.Query, inets, excludes []string) matcher {
	tbl := qry.Minion
	wheres := []gen.Condition{
		tbl.Inet.In(inets...),
		tbl.Inet.NotIn(excludes...),
		tbl.Status.Neq(uint8(model.MSDelete)),
	}

	return &inetModeMatch{
		qry:    qry,
		wheres: wheres,
	}
}

type inetModeMatch struct {
	qry    *query.Query
	wheres []gen.Condition
}

func (imm *inetModeMatch) Count(ctx context.Context) (int64, error) {
	dao := imm.qry.Minion.WithContext(ctx)
	return dao.Where(imm.wheres...).Count()
}

func (imm *inetModeMatch) Iter(ctx context.Context, batchSize int) iter.Seq2[[]*model.Minion, error] {
	dao := imm.qry.Minion.WithContext(ctx).Where(imm.wheres...)

	return func(yield func([]*model.Minion, error) bool) {
		var buf []*model.Minion
		_ = dao.FindInBatches(&buf, batchSize, func(tx gen.Dao, _ int) error {
			if !yield(buf, nil) {
				return sql.ErrNoRows
			}
			return nil
		})
	}
}
