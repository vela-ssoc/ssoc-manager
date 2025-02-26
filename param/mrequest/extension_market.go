package mrequest

import (
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/param/request"
)

type ExtensionMarketCreate struct {
	Name      string `json:"name"      validate:"required,lte=100"`
	Intro     string `json:"intro"     validate:"lte=1000"`
	Category  string `json:"category"  validate:"oneof=service task"` // service:服务插件 task:任务插件
	Content   string `json:"content"   validate:"required,lte=65535"`
	Changelog string `json:"changelog" validate:"lte=65535"`
}

type ExtensionMarketRecord struct {
	model.ExtensionMarket
	Records []*model.ExtensionRecord `json:"records"`
}

type ExtensionMarketPage struct {
	Page
	Category string `query:"category" validate:"omitempty,oneof=service task"`
}

type ExtensionMarketUpdate struct {
	request.Int64ID
	Intro     string `json:"intro"     validate:"lte=1000"`
	Content   string `json:"content"   validate:"required,lte=65535"`
	Changelog string `json:"changelog" validate:"lte=65535"`
}
