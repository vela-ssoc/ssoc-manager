package param

import (
	"errors"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
)

type NotifierCreate struct {
	Name      string   `json:"name"       validate:"required,lte=20"`
	Events    []string `json:"events"     validate:"lte=100,unique,dive,required"`
	Risks     []string `json:"risks"      validate:"lte=100,unique,dive,required"`
	EventCode []byte   `json:"event_code" validate:"lte=65535"`
	RiskCode  []byte   `json:"risk_code"  validate:"lte=65535"`
	Ways      []string `json:"ways"       validate:"lte=10,unique,dive,oneof=dong email wechat sms call"`
	Dong      string   `json:"dong"       validate:"omitempty,dong"`
	Email     string   `json:"email"      validate:"omitempty,email"`
	Mobile    string   `json:"mobile"     validate:"omitempty,mobile"`
}

func (na NotifierCreate) Validate() error {
	for _, way := range na.Ways {
		switch way {
		case "dong":
			if na.Dong == "" {
				return errors.New("选择咚咚通知时必须填写咚咚号")
			}
		case "email":
			if na.Email == "" {
				return errors.New("选择邮件通知时必须填写邮箱地址")
			}
		case "wechat", "sms", "call":
			if na.Mobile == "" {
				return errors.New("选择微信、短信或电话通知时必须填写手机号")
			}
		}
	}

	return nil
}

type NotifierUpdate struct {
	request.Int64ID
	NotifierCreate
}
