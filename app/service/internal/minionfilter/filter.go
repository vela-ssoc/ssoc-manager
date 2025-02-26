package minionfilter

import (
	"github.com/vela-ssoc/vela-common-mb/dal/dyncond"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"gorm.io/gen"
)

func New(qry *query.Query) {
}

type Filter struct {
	qry *query.Query
	tbl *dyncond.Tables
}

func (flt *Filter) Wheres() ([]gen.Condition, error) {
	return nil, nil
}
