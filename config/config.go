package config

import "gopkg.in/natefinch/lumberjack.v2"

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"database"`
	Logger   Logger   `json:"logger"`
}

type Server struct {
	Addr   string            `json:"addr"`
	Static map[string]string `json:"static"`
}

type Database struct {
	DSN         string   `json:"dsn"           validate:"required,lte=255"` // 数据库连接
	MaxOpenConn int      `json:"max_open_conn" validate:"gte=0"`            // 最大连接数
	MaxIdleConn int      `json:"max_idle_conn" validate:"gte=0"`            // 最大空闲连接数
	MaxLifeTime Duration `json:"max_life_time" validate:"gte=0"`            // 连接最大存活时长
	MaxIdleTime Duration `json:"max_idle_time" validate:"gte=0"`            // 空闲连接最大时长
}

type Logger struct {
	Console bool     `json:"console"`
	Level   LogLevel `json:"level"`
	*lumberjack.Logger
}

func (l Logger) Lumber() *lumberjack.Logger {
	if lum := l.Logger; lum != nil && lum.Filename != "" {
		return lum
	}

	return nil
}
