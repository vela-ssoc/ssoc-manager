package mservice

import (
	"log/slog"

	"github.com/vela-ssoc/vela-manager/applet/manager/mrequest"
)

func NewLog(logLevel, ormLevel *slog.LevelVar, log *slog.Logger) *Log {
	return &Log{
		logLevel: logLevel,
		ormLevel: ormLevel,
		log:      log,
	}
}

type Log struct {
	logLevel *slog.LevelVar
	ormLevel *slog.LevelVar
	log      *slog.Logger
}

func (lg *Log) Levels() (slog.Level, slog.Level) {
	lv := lg.logLevel.Level()
	ov := lg.ormLevel.Level()

	return lv, ov
}

func (lg *Log) Change(req *mrequest.LogChange) {
	if lvl := req.Log; lvl != "" {
		_ = lg.logLevel.UnmarshalText([]byte(lvl))
	}
	if lvl := req.ORM; lvl != "" {
		_ = lg.ormLevel.UnmarshalText([]byte(lvl))
	}
}
