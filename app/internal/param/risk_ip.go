package param

import (
	"time"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
)

type RiskIPCreate struct {
	IP       []string  `json:"ip"        validate:"gte=1,lte=100,dive,ip"` // IP 地址
	Kind     string    `json:"kind"      validate:"required,lte=20"`       // 风险类型
	Origin   string    `json:"origin"    validate:"lte=20"`                // 数据来源
	BeforeAt time.Time `json:"before_at"`                                  // 有效期
}

func (ria RiskIPCreate) Models() []*model.RiskIP {
	kind, origin, beforeAt := ria.Kind, ria.Origin, ria.BeforeAt
	ret := make([]*model.RiskIP, 0, len(ria.IP))
	for _, s := range ria.IP {
		ret = append(ret, &model.RiskIP{IP: s, Kind: kind, Origin: origin, BeforeAt: beforeAt})
	}

	return ret
}

type RiskIPUpdate struct {
	ID       int64     `json:"id,string" validate:"required"`
	IP       string    `json:"ip"        validate:"gte=1,lte=100,dive,ip"` // IP 地址
	Kind     string    `json:"kind"      validate:"required,lte=20"`       // 风险类型
	Origin   string    `json:"origin"    validate:"lte=20"`                // 数据来源
	BeforeAt time.Time `json:"before_at"`                                  // 有效期
}

type RiskIP struct {
	IP       string    `json:"ip"        validate:"ip"`              // IP 地址
	Kind     string    `json:"kind"      validate:"required,lte=20"` // 风险类型
	Origin   string    `json:"origin"    validate:"lte=20"`          // 数据来源
	BeforeAt time.Time `json:"before_at"`                            // 有效期
}

type RiskIPImport struct {
	Update bool      `json:"update"`                                       // 规则重复时执行更新操作，否则默认跳过
	Data   []*RiskIP `json:"data" validate:"gte=1,lte=1000,dive,required"` // 风险 IP 规则
}

func (rpi RiskIPImport) Models() []*model.RiskIP {
	size := len(rpi.Data)
	uniques := make(map[string]struct{}, size)
	ret := make([]*model.RiskIP, 0, size)

	for _, dat := range rpi.Data {
		unique := dat.IP + " " + dat.Kind
		if _, exist := uniques[unique]; exist {
			continue
		}

		uniques[unique] = struct{}{}
		ret = append(ret, &model.RiskIP{
			IP:       dat.IP,
			Kind:     dat.Kind,
			Origin:   dat.Origin,
			BeforeAt: dat.BeforeAt,
		})
	}

	return ret
}
