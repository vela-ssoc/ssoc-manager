package param

import "mime/multipart"

type ThirdCreate struct {
	Name string                `json:"name" query:"name" form:"name" validate:"filename"`
	Desc string                `json:"desc" query:"desc" form:"desc" validate:"lte=100"`
	File *multipart.FileHeader `json:"file" query:"file" form:"file" validate:"required"`
}

type ThirdUpdate struct {
	IntID
	Desc string                `json:"desc" query:"desc" form:"desc" validate:"lte=100"`
	File *multipart.FileHeader `json:"file" query:"file" form:"file"`
}
