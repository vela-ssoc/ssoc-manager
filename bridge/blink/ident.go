package blink

import (
	"encoding/json"
	"net"
	"time"

	"github.com/vela-ssoc/vela-common-mba/encipher"
)

// Ident broker 节点的认证信息
type Ident struct {
	ID         int64     `json:"id"`         // ID
	Secret     string    `json:"secret"`     // 密钥
	Inet       net.IP    `json:"inet"`       // IPv4 地址
	MAC        string    `json:"mac"`        // MAC 地址
	Semver     string    `json:"semver"`     // 版本
	Goos       string    `json:"goos"`       // runtime.GOOS
	Arch       string    `json:"arch"`       // runtime.GOARCH
	CPU        int       `json:"cpu"`        // runtime.NumCPU
	PID        int       `json:"pid"`        // os.Getpid
	Workdir    string    `json:"workdir"`    // os.Getwd
	Executable string    `json:"executable"` // os.Executable
	Username   string    `json:"username"`   // user.Current
	Hostname   string    `json:"hostname"`   // os.Hostname
	TimeAt     time.Time `json:"time_at"`    // 发起时间
}

// String fmt.Stringer
func (ide Ident) String() string {
	dat, _ := json.MarshalIndent(ide, "", "    ")
	return string(dat)
}

// decrypt 将数据解密至 Ident
func (ide *Ident) decrypt(enc []byte) error {
	return encipher.DecryptJSON(enc, ide)
}
