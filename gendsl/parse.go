package gendsl

import (
	"gorm.io/gorm"
)

func Parse(db *gorm.DB, models []any) (*Table, error) {
	return nil, nil
}

type Table struct{}

func (tbl *Table) Select() {
}
