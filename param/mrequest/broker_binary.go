package mrequest

import "github.com/vela-ssoc/vela-common-mb/param/request"

type BrokerDownload struct {
	request.Int64ID
	BrokerID int64 `query:"broker_id" validate:"required"`
}

type BrokerBinaryLatest struct {
	Goos string `json:"goos" query:"goos" validate:"required,lte=10"`
	Arch string `json:"arch" query:"arch" validate:"required,lte=10"`
}
