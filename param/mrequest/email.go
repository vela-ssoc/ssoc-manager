package mrequest

import "github.com/vela-ssoc/vela-common-mb/param/request"

type EmailCreate struct {
	Host     string `json:"host"     validate:"hostname_rfc1123"`
	Username string `json:"username" validate:"email,lte=50"`
	Password string `json:"password" validate:"required,lte=50"`
	Enable   bool   `json:"enable"`
}

type EmailUpdate struct {
	request.Int64ID
	EmailCreate
}
