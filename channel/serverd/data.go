package serverd

import (
	"time"

	"github.com/vela-ssoc/ssoc-common/data/datatype"
)

type authRequest struct {
	Secret datatype.Ciphertext `json:"secret"` // broker 密钥
	Semver string              `json:"semver"` // broker 版本号
}

type authResponse struct {
	Code     int      `json:"code"`
	Message  string   `json:"message"`
	Database Database `json:"database"`
}

type Database struct {
	DSN         datatype.Ciphertext `json:"dsn"`
	MaxOpenConn int                 `json:"max_open_conn"`
	MaxIdleConn int                 `json:"max_idle_conn"`
	MaxLifeTime time.Duration       `json:"max_life_time"`
	MaxIdleTime time.Duration       `json:"max_idle_time"`
}
