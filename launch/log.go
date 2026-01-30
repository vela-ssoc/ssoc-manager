package launch

import (
	"log/slog"
	"os"

	"github.com/vela-ssoc/ssoc-common/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)

func initLogHandler() logger.Handler {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	file := &lumberjack.Logger{
		Filename:   "resources/log/application.jsonl",
		MaxSize:    100,
		MaxBackups: 10,
		LocalTime:  true,
		Compress:   true,
	}

	return logger.NewMultiHandler(
		logger.NewTint(os.Stdout, opts),
		slog.NewJSONHandler(file, opts),
	)
}
