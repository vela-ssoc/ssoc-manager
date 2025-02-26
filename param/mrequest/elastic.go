package mrequest

import "github.com/vela-ssoc/vela-common-mb/param/request"

type ElasticCreate struct {
	// Host     string   `json:"host"     validate:"http"`                          // ES 地址
	Username string   `json:"username"`                                          // ES 用户名
	Password string   `json:"password"`                                          // ES 密码
	Hosts    []string `json:"hosts"    validate:"gte=1,lte=20,unique,dive,http"` // ES 地址
	Desc     string   `json:"desc"     validate:"lte=100"`                       // 简介
	Enable   bool     `json:"enable"`                                            // 是否选中
}

type ElasticUpdate struct {
	request.Int64ID
	ElasticCreate
}

type ElasticDetect struct {
	Host     string `json:"host"     validate:"http"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ElasticDetects []*elasticDetect

func (eds ElasticDetects) Addrs() []string {
	ret := make([]string, 0, len(eds))
	for _, ed := range eds {
		ret = append(ret, ed.HTTP)
	}
	return ret
}

type elasticDetect struct {
	HTTP string `json:"http"`
}
