package profile

import (
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Server   Server   `json:"server"`   // HTTP 服务
	Database Database `json:"database"` // 数据库
	Logger   Logger   `json:"logger"`   // 日志
}

type Server struct {
	Addr   string   `json:"addr"`                            // 监听地址
	Cert   string   `json:"cert"`                            // 证书
	Pkey   string   `json:"pkey"`                            // 私钥
	Static string   `json:"static"`                          // 静态资源路径
	Vhosts []string `json:"vhosts" validate:"dive,required"` // 虚拟主机头
}

type Database struct {
	DSN         string   `json:"dsn"           validate:"required,lte=255"` // 数据库连接
	Migrate     bool     `json:"migrate"`                                   // 是否合并差异
	MaxOpenConn int      `json:"max_open_conn" validate:"gte=0"`            // 最大连接数
	MaxIdleConn int      `json:"max_idle_conn" validate:"gte=0"`            // 最大空闲连接数
	MaxLifeTime duration `json:"max_life_time"`                             // 连接最大存活时长
	MaxIdleTime duration `json:"max_idle_time"`                             // 空闲连接最大时长
}

type Logger struct {
	*lumberjack.Logger
	Console bool   `json:"console"`
	Level   string `json:"level"   validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
}

type duration time.Duration

func (du *duration) UnmarshalText(bs []byte) error {
	pd, err := time.ParseDuration(string(bs))
	if err == nil {
		*du = duration(pd)
	}

	return err
}
