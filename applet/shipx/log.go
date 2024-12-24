package shipx

import (
	"fmt"
	"log/slog"

	"github.com/xgfone/ship/v5"
)

func NewLog(log *slog.Logger) ship.Logger {
	return &shipLog{log: log}
}

type shipLog struct {
	log *slog.Logger
}

func (s *shipLog) Tracef(format string, args ...any) {
	s.logf(slog.LevelDebug, format, args...)
}

func (s *shipLog) Debugf(format string, args ...any) {
	s.logf(slog.LevelDebug, format, args...)
}

func (s *shipLog) Infof(format string, args ...any) {
	s.logf(slog.LevelInfo, format, args...)
}

func (s *shipLog) Warnf(format string, args ...any) {
	s.logf(slog.LevelWarn, format, args...)
}

func (s *shipLog) Errorf(format string, args ...any) {
	s.logf(slog.LevelError, format, args...)
}

func (s *shipLog) logf(level slog.Level, format string, args ...any) {
	if !s.log.Enabled(nil, level) {
		return
	}

	size := len(args)
	if size == 0 {
		s.log.Log(nil, level, format)
		return
	}

	var not bool
	attrs := make([]slog.Attr, 0, size)
	for _, arg := range args {
		if attr, ok := arg.(slog.Attr); ok {
			attrs = append(attrs, attr)
		} else {
			not = true
			break
		}
	}
	if not {
		msg := fmt.Sprintf(format, args...)
		s.log.Log(nil, level, msg)
	} else {
		s.log.LogAttrs(nil, level, format, attrs...)
	}
}
