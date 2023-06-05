package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/vela-ssoc/vela-common-mb/logback"
	"github.com/vela-ssoc/vela-manager/infra/banner"
	"github.com/vela-ssoc/vela-manager/launch"
)

func main() {
	flagSet := flag.NewFlagSet("安全平台", flag.ExitOnError)
	printExit := flagSet.Bool("v", false, "打印版本号就退出")
	configPath := flagSet.String("c", "resources/config/manager.yaml", "配置文件路径")
	_ = flagSet.Parse(os.Args[1:])

	if banner.WriteTo(os.Stdout); *printExit {
		return
	}

	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()

	slog := logback.Stdout()
	slog.Info("按 Ctrl+C 结束运行")

	if err := launch.Run(ctx, *configPath, slog); err != nil {
		slog.Warnf("程序启动错误：%v", err)
	}
	slog.Warn("程序已停止运行")
}
