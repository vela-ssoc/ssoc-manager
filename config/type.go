package config

import (
	"log/slog"
	"time"
)

type LogLevel string

func (l LogLevel) LevelVar() *slog.LevelVar {
	lvl := new(slog.LevelVar)
	_ = lvl.UnmarshalText([]byte(l))

	return lvl
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
