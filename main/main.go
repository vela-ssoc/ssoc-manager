package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vela-ssoc/ssoc-manager/banner"
	"github.com/vela-ssoc/ssoc-manager/launch"
)

func main() {
	args := os.Args
	set := flag.NewFlagSet(args[0], flag.ExitOnError)
	v := set.Bool("v", false, "打印版本并退出")
	c := set.String("c", "resources/config/application.jsonc", "配置文件路径")
	_ = set.Parse(args[1:])

	if banner.ANSI(os.Stdout); *v {
		return
	}

	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()

	logOpt := &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	log := slog.New(slog.NewJSONHandler(os.Stdout, logOpt))
	log.Info("按 [Ctrl+C] 停止运行")
	if err := launch.Run(ctx, *c); err != nil {
		log.Error("程序运行错误", slog.Any("error", err))
	} else {
		log.Warn("程序运行结束")
	}
}
