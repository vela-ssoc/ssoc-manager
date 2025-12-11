package serverd

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common/data/datatype"
)

type authRequest struct {
	Secret datatype.Ciphertext `json:"secret" validate:"required"` // broker 密钥
	Semver string              `json:"semver" validate:"required"` // broker 版本号
	Goos   string              `json:"goos"`
	Goarch string              `json:"goarch"`
}

type authResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	BootConfig *BootConfig `json:"boot_config"`
}

type BootConfig struct {
	DSN         string        `json:"dsn"`
	MaxOpenConn int           `json:"max_open_conn"`
	MaxIdleConn int           `json:"max_idle_conn"`
	MaxLifeTime time.Duration `json:"max_life_time"`
	MaxIdleTime time.Duration `json:"max_idle_time"`
}

type BootConfigurer interface {
	BootConfig(ctx context.Context) (*BootConfig, error)
}
