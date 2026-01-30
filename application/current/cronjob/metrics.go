package cronjob

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/robfig/cron/v3"
	"github.com/vela-ssoc/ssoc-common/cronv3"
)

type MetricsConfigFunc func(ctx context.Context) (pushURL string, opts *metrics.PushOptions, err error)

func NewMetrics(cfg MetricsConfigFunc) cronv3.Tasker {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	hostname, _ := os.Hostname()

	label := fmt.Sprintf(`instance="ssoc-manager",instance_type="manager",instance_name="%s",goos="%s",goarch="%s"`, hostname, goos, goarch)

	return &metricsJob{
		cfg:   cfg,
		label: label,
	}
}

type metricsJob struct {
	cfg   MetricsConfigFunc
	label string
}

func (*metricsJob) Info() cronv3.TaskInfo {
	return cronv3.TaskInfo{
		Name:      "上报系统指标",
		Timeout:   9 * time.Second,
		CronSched: cron.Every(10 * time.Second),
	}
}

func (m *metricsJob) Call(ctx context.Context) error {
	pushURL, opts, err := m.cfg(ctx)
	if err != nil {
		return err
	}
	opts.ExtraLabels = m.label

	return metrics.PushMetricsExt(ctx, pushURL, m.defaultWrite, opts)
}

func (m *metricsJob) defaultWrite(w io.Writer) {
	metrics.WritePrometheus(w, true)
	metrics.WriteFDMetrics(w)
}
