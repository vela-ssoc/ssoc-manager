package profile

import (
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

type Config struct {
	Active   string   `json:"active"`   // 多环境配置
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
	DSN         string   `json:"dsn"           validate:"required,lte=255"`                       // 数据库连接
	Level       string   `json:"level"         validate:"omitempty,oneof=INFO WARN ERROR SILENT"` // 日志输出级别
	Migrate     bool     `json:"migrate"`                                                         // 是否合并差异
	MaxOpenConn int      `json:"max_open_conn" validate:"gte=0"`                                  // 最大连接数
	MaxIdleConn int      `json:"max_idle_conn" validate:"gte=0"`                                  // 最大空闲连接数
	MaxLifeTime duration `json:"max_life_time"`                                                   // 连接最大存活时长
	MaxIdleTime duration `json:"max_idle_time"`                                                   // 空闲连接最大时长
}

func (c Config) GormLevel() logger.LogLevel {
	lvl := c.Database.Level
	if lvl == "" {
		lvl = c.Logger.Level
	}

	switch strings.ToUpper(lvl) {
	case "SILENT", "ERROR":
		return logger.Error
	case "WARN":
		return logger.Warn
	default:
		return logger.Info
	}
}

type duration time.Duration

func (du *duration) UnmarshalText(bs []byte) error {
	pd, err := time.ParseDuration(string(bs))
	if err == nil {
		*du = duration(pd)
	}

	return err
}

func (du *duration) String() string {
	return time.Duration(*du).String()
}

func (du *duration) Duration() time.Duration {
	return time.Duration(*du)
}
