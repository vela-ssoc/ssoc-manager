package profile

import (
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

// Config 全局配置文件。
type Config struct {
	Active   string   `json:"active"`   // 多环境配置
	Server   Server   `json:"server"`   // HTTP 服务
	Database Database `json:"database"` // 数据库
	Logger   Logger   `json:"logger"`   // 日志
}

type Server struct {
	Addr   string   `json:"addr"`                            // 监听地址
	Cert   string   `json:"cert"   validate:"lte=500"`       // 证书
	Pkey   string   `json:"pkey"   validate:"lte=500"`       // 私钥
	Static string   `json:"static" validate:"lte=500"`       // 静态资源路径
	Vhosts []string `json:"vhosts" validate:"dive,required"` // 虚拟主机头
}

type Database struct {
	DSN         string   `json:"dsn"           validate:"required,lte=255"`                      // 数据库连接
	Level       string   `json:"level"         validate:"omitempty,oneof=DEBUG INFO WARN ERROR"` // 日志输出级别
	Migrate     bool     `json:"migrate"`                                                        // 是否合并差异
	MaxOpenConn int      `json:"max_open_conn" validate:"gte=0"`                                 // 最大连接数
	MaxIdleConn int      `json:"max_idle_conn" validate:"gte=0"`                                 // 最大空闲连接数
	MaxLifeTime duration `json:"max_life_time" validate:"gte=0"`                                 // 连接最大存活时长
	MaxIdleTime duration `json:"max_idle_time" validate:"gte=0"`                                 // 空闲连接最大时长
}

func (c Config) GormLevel() logger.LogLevel {
	lvl := c.Database.Level
	if lvl == "" {
		lvl = c.Logger.Level
	}

	switch strings.ToUpper(lvl) {
	case "ERROR":
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
