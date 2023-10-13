package param

import (
	"image"

	"github.com/vela-ssoc/vela-manager/app/totp"
)

type AuthBase struct {
	Username string `json:"username" validate:"required,lte=20"`
	Password string `json:"password" validate:"required,gte=6,lte=32"`
}

type AuthPicture struct {
	ID    string `json:"id"`
	Board string `json:"board"`
	Thumb string `json:"thumb"`
}

type AuthVerify struct {
	AuthBase
	ID     string        `json:"id" validate:"required,lte=100"`
	Points picturePoints `json:"points"     validate:"gte=1,lte=6,dive"`
}

type AuthLogin struct {
	AuthBase
	CaptchaID string `json:"captcha_id" validate:"required,lte=255"`
	Code      string `json:"code"`
}

type AuthDong struct {
	AuthBase
	CaptchaID string `json:"captcha_id" validate:"required,lte=100"`
}

type AuthNeedDong struct {
	Ding bool `json:"ding"`
}

type picturePoint struct {
	X int `json:"x" validate:"gte=0,lte=10000"`
	Y int `json:"y" validate:"gte=0,lte=10000"`
}

type picturePoints []*picturePoint

func (pps picturePoints) Convert() []*image.Point {
	ret := make([]*image.Point, 0, len(pps))
	for _, pp := range pps {
		pt := (*image.Point)(pp) // 内存布局一样可以强制类型转换
		ret = append(ret, pt)
	}

	return ret
}

type AuthSubmit struct {
	UID  string `json:"uid"  validate:"required,gte=255"`
	Code string `json:"code" validate:"len=6,numeric"`
}

type AuthUID struct {
	UID string `json:"uid"  validate:"required,gte=255"`
}

type AuthTotpResp struct {
	*totp.TOTP
	URL string `json:"url"`
}
