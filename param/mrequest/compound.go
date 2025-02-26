package mrequest

import (
	"time"

	"github.com/vela-ssoc/vela-common-mb/param/request"
)

type CompoundCreate struct {
	Name       string         `json:"name"       validate:"required,lte=50"`
	Substances request.Int64s `json:"substances" validate:"gte=1,lte=100,unique"`
	Desc       string         `json:"desc"       validate:"omitempty,lte=255"`
}

type CompoundUpdate struct {
	CompoundCreate
	ID      int64 `json:"id,string" validate:"required"`
	Version int64 `json:"version"`
}

type CompoundVO struct {
	ID         int64           `json:"id,string"  gorm:"column:id"`
	Name       string          `json:"name"       gorm:"column:name"`
	Desc       string          `json:"desc"       gorm:"column:desc"`
	Substances request.IDNames `json:"substances" gorm:"column:substances"`
	Version    int64           `json:"version"    gorm:"column:version"`
	CreatedAt  time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time       `json:"updated_at" gorm:"column:updated_at"`
}
