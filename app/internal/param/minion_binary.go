package param

import (
	"mime/multipart"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/param/request"
)

type NodeBinaryCreate struct {
	Name       string                `json:"name"       form:"name"       validate:"required,lte=100"`
	Semver     model.Semver          `json:"semver"     form:"semver"     validate:"semver"`
	Goos       string                `json:"goos"       form:"goos"       validate:"oneof=linux windows darwin"`
	Arch       string                `json:"arch"       form:"arch"       validate:"oneof=amd64 386 arm64 arm loong64 riscv64"`
	Unstable   bool                  `json:"unstable"   form:"unstable"`
	Customized string                `json:"customized" form:"customized" validate:"lte=255"`
	Changelog  string                `json:"changelog"  form:"changelog"  validate:"lte=65500"`
	Caution    string                `json:"caution"    form:"caution"    validate:"lte=65500"`
	Ability    string                `json:"ability"    form:"ability"    validate:"lte=65500"`
	File       *multipart.FileHeader `json:"file"       form:"file"       validate:"required"`
}

type MinionBinaryClassify struct {
	Structures []*MinionBinaryStructure `json:"structures"`
	Customized string                   `json:"customized"`
}

type MinionBinaryStructure struct {
	Goos string `json:"goos" gorm:"column:goos"`
	Arch string `json:"arch" gorm:"column:arch"`
}

type MinionBinaryUpdate struct {
	request.Int64ID
	Changelog string `json:"changelog" validate:"lte=65500"`
	Caution   string `json:"caution"   validate:"lte=65500"`
	Ability   string `json:"ability"   validate:"lte=65500"`
}
