package config

import (
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*lumberjack.Logger
	Console bool   `json:"console"`
	Level   string `json:"level"   validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
}

func (l Logger) Lumber() *lumberjack.Logger {
	if lum := l.Logger; lum != nil && lum.Filename != "" {
		return lum
	}

	return nil
}

const DefaultLogFilename = "resources/log/application.jsonl"

type Config struct {
	Database Database `json:"database"`
}

type Database struct {
	URI string `json:"uri" validate:"required"`
	DSN string `json:"dsn" validate:"required"`
}
