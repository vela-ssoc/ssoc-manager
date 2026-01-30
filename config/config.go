package config

import (
	"log/slog"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Filename 默认配置文件路径。
const Filename = "resources/config/application.jsonc"

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"database"`
	Logger   Logger   `json:"logger"`
}

type Server struct {
	Addr string `json:"addr"`
}

type Database struct {
	DSN         string   `json:"dsn"           validate:"required,lte=255"` // 数据库连接
	MaxOpenConn int      `json:"max_open_conn" validate:"gte=0"`            // 最大连接数
	MaxIdleConn int      `json:"max_idle_conn" validate:"gte=0"`            // 最大空闲连接数
	MaxLifeTime Duration `json:"max_life_time" validate:"gte=0"`            // 连接最大存活时长
	MaxIdleTime Duration `json:"max_idle_time" validate:"gte=0"`            // 空闲连接最大时长
}

type Logger struct {
	Console bool           `json:"console"`
	Level   *slog.LevelVar `json:"level"`
	*lumberjack.Logger
}

func (l *Logger) Lumber() *lumberjack.Logger {
	if lb := l.Logger; lb != nil && lb.Filename != "" {
		return lb
	}

	return nil
}

type Duration time.Duration

func (d *Duration) UnmarshalText(b []byte) error {
	du, err := time.ParseDuration(string(b))
	if err == nil {
		*d = Duration(du)
	}

	return err
}

func (d Duration) MarshalText() ([]byte, error) {
	s := time.Duration(d).String()
	return []byte(s), nil
}
