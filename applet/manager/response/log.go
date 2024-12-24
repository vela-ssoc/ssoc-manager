package response

import "log/slog"

type LogLevel struct {
	Log slog.Level `json:"log"`
	ORM slog.Level `json:"orm"`
}
