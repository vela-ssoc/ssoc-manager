package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/vela-ssoc/ssoc-common/banner"
	"github.com/vela-ssoc/ssoc-manager/launch"
)

func main() {
	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	v := set.Bool("v", false, "打印版本")
	c := set.String("c", "resources/config/application.jsonc", "配置文件")
	_ = set.Parse(os.Args[1:])
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	slog.Info("按 Ctrl+C 停止运行")
	err := launch.Run(ctx, *c)
	cause := context.Cause(ctx)
	slog.Warn("程序停止运行", "error", err, "cause", cause)
}
