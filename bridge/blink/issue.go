package blink

import (
	"encoding/json"

	"github.com/vela-ssoc/vela-common-mba/encipher"
)

// Issue broker 节点认证成功后返回的信息
type Issue struct {
	Name   string `json:"name"`   // broker 名字
	Passwd []byte `json:"passwd"` // 通信加解密密钥
	// Listen   Listen        `json:"listen"`   // 服务监听配置
	// Logger   conf.Logger   `json:"logger"`   // 日志配置
	// Database conf.Database `json:"database"` // 数据库配置
}

// String fmt.Stringer
func (iss Issue) String() string {
	dat, _ := json.MarshalIndent(iss, "", "    ")
	return string(dat)
}

// encrypt 将 Issue 加密
func (iss Issue) encrypt() ([]byte, error) {
	return encipher.EncryptJSON(iss)
}
