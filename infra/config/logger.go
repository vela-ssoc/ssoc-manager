package config

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志配置
type Logger struct {
	Level     string `json:"level"     yaml:"level"`     // 日志级别
	Console   bool   `json:"console"   yaml:"console"`   // 是否输出到控制台
	Colorful  bool   `json:"colorful"  yaml:"colorful"`  // 输出日志带颜色
	Directory string `json:"directory" yaml:"directory"` // 日志文件存放目录
	Maxsize   int    `json:"maxsize"   yaml:"maxsize"`   // 文件分割大小
	MaxAge    int    `json:"maxage"    yaml:"maxage"`    // 文件留存天数
	Backup    int    `json:"backup"    yaml:"backup"`    // 分割文件留存个数
	Localtime bool   `json:"localtime" yaml:"localtime"` // 分割备份文件名是否使用本地时区
	Compress  bool   `json:"compress"  yaml:"compress"`  // 分割备份文件是否压缩
}

// Zap 根据创建 zap 日志
func (l Logger) Zap() *zap.Logger {
	console := l.Console
	var filename string
	if dir := l.Directory; dir != "" {
		filename = filepath.Join(dir, "manager.log")
	}
	// 既不输出到控制台又不输出到日志文件
	if !console && filename == "" {
		return zap.NewNop()
	}

	prod := zap.NewProductionEncoderConfig()
	prod.EncodeTime = zapcore.ISO8601TimeEncoder
	if l.Colorful {
		prod.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		prod.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	var syncer zapcore.WriteSyncer
	if console {
		syncer = zapcore.AddSync(os.Stdout)
	}
	if filename != "" {
		lumber := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    l.Maxsize,
			MaxAge:     l.MaxAge,
			MaxBackups: l.Backup,
			LocalTime:  l.Localtime,
			Compress:   l.Compress,
		}
		if syncer == nil {
			syncer = zapcore.AddSync(lumber)
		} else {
			syncer = zapcore.NewMultiWriteSyncer(syncer, zapcore.AddSync(lumber))
		}
	}

	encoder := zapcore.NewConsoleEncoder(prod)
	level := zapcore.WarnLevel
	_ = level.Set(l.Level) // 就算设置失败还是默认值 WarnLevel
	core := zapcore.NewCore(encoder, syncer, level)

	return zap.New(core, zap.WithCaller(true), zap.AddStacktrace(zapcore.ErrorLevel))
}
