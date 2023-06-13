package param

import (
	"mime/multipart"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
)

type NodeBinaryCreate struct {
	Name      string                `json:"name"      form:"name"      validate:"required,lte=100"`
	Semver    model.Semver          `json:"semver"    form:"semver"    validate:"semver"`
	Goos      string                `json:"goos"      form:"goos"      validate:"oneof=linux windows darwin"`
	Arch      string                `json:"arch"      form:"arch"      validate:"oneof=amd64 386 arm64 arm"`
	Changelog string                `json:"changelog" form:"changelog" validate:"lte=2048"`
	File      *multipart.FileHeader `json:"file"      form:"file"      validate:"required"`
}
