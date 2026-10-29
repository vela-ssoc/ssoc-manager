package cronfunc

import (
	"context"
	"io"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/robfig/cron/v3"
	"github.com/vela-ssoc/ssoc-common-mb/crontab"
	"github.com/vela-ssoc/ssoc-common-mb/vmexport"
)

type VMetric struct {
	cfg    vmexport.VictoriaPushConfigLoader
	writes []vmexport.MetricWriter
}

func NewVMetric(cfg vmexport.VictoriaPushConfigLoader, writes []vmexport.MetricWriter) *VMetric {
	return &VMetric{cfg: cfg, writes: writes}
}

func (vm *VMetric) Options() crontab.JobOptions {
	return crontab.JobOptions{
		ID:      "86eb816810333dee1049300083b1e9a5b34f058a",
		Name:    "上报系统监控指标",
		Sched:   cron.Every(10 * time.Second),
		Timeout: 8 * time.Second,
	}
}

func (vm *VMetric) Execute(ctx context.Context) error {
	pushURL, opts, err := vm.cfg.LoadVictoriaPushConfig(ctx)
	if err != nil {
		return err
	}

	return metrics.PushMetricsExt(ctx, pushURL, vm.writeWithContext(ctx), opts)
}

func (vm *VMetric) writeWithContext(ctx context.Context) func(io.Writer) {
	return func(w io.Writer) {
		metrics.WritePrometheus(w, true)
		metrics.WriteFDMetrics(w)

		for _, mw := range vm.writes {
			mw.WriteMetric(ctx, w)
		}
	}
}
