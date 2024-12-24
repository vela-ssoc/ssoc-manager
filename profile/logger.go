package profile

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*lumberjack.Logger
	Console bool   `json:"console"`
	Level   string `json:"level"   validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
}

func (l *Logger) Writer() io.WriteCloser {
	var writers []io.Writer
	var closer io.Closer
	var writer io.Writer

	if l.Console {
		writers = append(writers, os.Stdout)
	}
	if lumber := l.Logger; lumber != nil && lumber.Filename != "" {
		writers = append(writers, lumber)
		closer = lumber
	}

	cnt := len(writers)
	if cnt == 0 {
		writer = os.Stdout
	} else if cnt == 1 {
		writer = writers[0]
	} else {
		writer = io.MultiWriter(writers...)
	}

	return &logWriter{
		w: writer,
		c: closer,
	}
}

type logWriter struct {
	w io.Writer
	c io.Closer
}

func (lw *logWriter) Write(p []byte) (int, error) {
	return lw.w.Write(p)
}

func (lw *logWriter) Close() error {
	if c := lw.c; c != nil {
		return c.Close()
	}
	return nil
}
