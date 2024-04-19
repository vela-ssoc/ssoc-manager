package param

import (
	"mime/multipart"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
)

type ThirdCreate struct {
	Name       string                `json:"name"       query:"name"       form:"name"       validate:"filename"`
	Desc       string                `json:"desc"       query:"desc"       form:"desc"       validate:"lte=100"`
	Customized string                `json:"customized" query:"customized" form:"customized" validate:"lte=50"`
	File       *multipart.FileHeader `json:"file"       query:"file"       form:"file"       validate:"required"`
}

type ThirdUpdate struct {
	IntID
	Customized string                `json:"customized" query:"customized" form:"customized" validate:"lte=50"`
	Desc       string                `json:"desc"       query:"desc"       form:"desc"       validate:"lte=100"`
	File       *multipart.FileHeader `json:"file"       query:"file"       form:"file"`
}

type ThirdListItem struct {
	model.ThirdCustomized
	Records []*model.Third `json:"records"`
}
