package mrequest

import "github.com/vela-ssoc/vela-common-mb/param/request"

type BrokerDownload struct {
	request.Int64ID
	BrokerID int64 `query:"broker_id" validate:"required"`
}
