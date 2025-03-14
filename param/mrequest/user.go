package mrequest

import (
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
)

type UserSummaries []*userSummary

type userSummary struct {
	ID        int64  `json:"id,string"  gorm:"column:id"`
	Username  string `json:"username"   gorm:"column:username"`
	Nickname  string `json:"nickname"   gorm:"column:nickname"`
	Dong      string `json:"dong"       gorm:"column:dong"`
	Enable    bool   `json:"enable"     gorm:"column:enable"`
	AccessKey string `json:"access_key" gorm:"column:access_key"`
}

type UserSudo struct {
	request.Int64ID
	Nickname string `json:"nickname"  validate:"required,lte=20"`
	Enable   bool   `json:"enable"`
	Password string `json:"password"  validate:"omitempty,password"`
}

type UserCreate struct {
	Username string           `json:"username" validate:"username"`                                // 注册用户名
	Nickname string           `json:"nickname" validate:"gte=2,lte=20"`                            // 昵称
	Domain   model.UserDomain `json:"domain"   validate:"oneof=1 2"`                               // 账号类型
	Password string           `json:"password" validate:"required_if=Domain 1,omitempty,password"` // 密码
	Enable   bool             `json:"enable"`
}

type UserPasswd struct {
	Original string `json:"original" validate:"required,lte=32"`
	Password string `json:"password" validate:"password"`
}
