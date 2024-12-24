package service

import (
	"log/slog"

	"github.com/vela-ssoc/vela-manager/applet/manager/request"
)

func NewLog(logLevel, ormLevel *slog.LevelVar) *Log {
	return &Log{
		logLevel: logLevel,
		ormLevel: ormLevel,
	}
}

type Log struct {
	logLevel *slog.LevelVar
	ormLevel *slog.LevelVar
}

func (lg *Log) Levels() (slog.Level, slog.Level) {
	lv := lg.logLevel.Level()
	ov := lg.ormLevel.Level()

	return lv, ov
}

func (lg *Log) Change(req *request.LogChange) {
	if lvl := req.Log; lvl != "" {
		_ = lg.logLevel.UnmarshalText([]byte(lvl))
	}
	if lvl := req.ORM; lvl != "" {
		_ = lg.ormLevel.UnmarshalText([]byte(lvl))
	}
}
