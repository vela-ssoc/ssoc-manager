package config

import "github.com/vela-ssoc/vela-common-mb/dbms"

// Config 配置参数
type Config struct {
	Server   Server      `json:"server"   yaml:"server"`   // HTTP 服务配置
	Database dbms.Config `json:"database" yaml:"database"` // 数据库配置
	Logger   Logger      `json:"logger"   yaml:"logger"`   // 日志配置
	Section  Section     `json:"section"  yaml:"section"`  // 其他信息
	SIEM     SIEM        `json:"siem"     yaml:"siem"`
}

type SIEM struct {
	URL   string `json:"url"   yaml:"url"`
	Token string `json:"token" yaml:"token"`
}
