package mrequest

import "github.com/vela-ssoc/ssoc-common-mb/param/request"

type ViewHTML struct {
	request.Int64ID
	Secret string `json:"secret" query:"secret" validate:"required,lte=255"`
}
