package mrequest

import "github.com/vela-ssoc/vela-common-mb/param/request"

type CertCreate struct {
	Name        string `json:"name"        validate:"required,lte=50"`
	Certificate string `json:"certificate" validate:"required,lte=65535"`
	PrivateKey  string `json:"private_key" validate:"required,lte=65535"`
}

type CertUpdate struct {
	request.Int64ID
	CertCreate
}
