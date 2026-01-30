package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxaccept"
)

type BootConfig struct {
	cfg config.Database
}

func NewBootConfig(cfg config.Database) *BootConfig {
	return &BootConfig{cfg: cfg}
}

func (bc *BootConfig) LoadBoot(context.Context) (*muxaccept.BrokConfig, error) {
	c := bc.cfg
	return &muxaccept.BrokConfig{
		DSN:         c.DSN,
		MaxOpenConn: c.MaxOpenConn,
		MaxIdleConn: c.MaxIdleConn,
		MaxLifeTime: time.Duration(c.MaxLifeTime),
		MaxIdleTime: time.Duration(c.MaxIdleTime),
	}, nil
}
