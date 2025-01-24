package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vela-ssoc/vela-common-mb/logback"
	"github.com/vela-ssoc/vela-manager/banner"
	"github.com/vela-ssoc/vela-manager/launch"
)

func main() {
	var version bool
	var config string

	args := os.Args
	fset := flag.NewFlagSet(args[0], flag.ExitOnError)
	fset.BoolVar(&version, "v", false, "打印版本号并退出")
	fset.StringVar(&config, "c", "resources/config/manager.yaml", "配置文件路径")
	_ = fset.Parse(args[1:])

	if banner.ANSI(os.Stdout); version {
		return
	}

	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()

	log := slog.Default()
	log.Info("按 [Ctrl+C] 停止运行")
	if err := launch.Run(ctx, config, logback.Stdout()); err != nil {
		log.Error("程序运行错误", slog.Any("error", err))
	} else {
		log.Warn("程序运行结束")
	}
}

func mainV1() {
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

	log := slog.Default()
	log.Info("按 [Ctrl+C] 停止运行")
	if err := launch.Run1(ctx, *c); err != nil {
		log.Error("程序运行错误", slog.Any("error", err))
	} else {
		log.Warn("程序运行结束")
	}
}
