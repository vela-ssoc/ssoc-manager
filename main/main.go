package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/vela-ssoc/ssoc-common/banner"
	"github.com/vela-ssoc/ssoc-manager/launch"
)

func main() {
	args := os.Args
	set := flag.NewFlagSet(args[0], flag.ExitOnError)
	v := set.Bool("v", false, "打印版本并退出")
	c := set.String("c",
		"resources/config/application.jsonc", "配置文件路径")
	_ = set.Parse(args[1:])

	if _, _ = banner.ANSI(os.Stdout); *v {
		return
	}

	// https://github.com/golang/go/issues/67182
	for _, fp := range []string{"resources/.crash.txt", ".crash.txt"} {
		if f, _ := os.Create(fp); f != nil {
			_ = debug.SetCrashOutput(f, debug.CrashOptions{})
			_ = f.Close()
			break
		}
	}

	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()

	slog.Info("按 [Ctrl+C] 停止运行")
	if err := launch.Run(ctx, *c); err != nil {
		slog.Error("程序运行错误", slog.Any("error", err))
	} else {
		slog.Warn("程序运行结束")
	}
}

// POST /api/v1/agent/task/stop
