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
	URI string `json:"uri" validate:"mongodb_connection_string"`
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
