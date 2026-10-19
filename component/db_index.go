package component

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
)

func CreateIndex(ctx context.Context, db polydb.Sourcer) error {
	return db.MDB().CreateIndex(ctx)
}
