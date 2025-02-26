package mrequest

import (
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/param/request"
)

type BrokerCreate struct {
	Name       string   `json:"name"           validate:"lte=20"`                                     // 名字只是为了有辨识度
	LAN        []string `json:"lan"            validate:"omitempty,unique,lte=10,dive,hostname_port"` // 内部连接地址
	VIP        []string `json:"vip"            validate:"omitempty,unique,lte=10,dive,hostname_port"` // 外部连接地址
	Bind       string   `json:"bind"           validate:"required,lte=22,hostname_port"`              // 监听地址
	Servername string   `json:"servername"     validate:"omitempty,lte=255,hostname_rfc1123"`         // TLS 证书校验用
	CertID     int64    `json:"cert_id,string"`                                                       // 关联证书 ID
}

type BrokerUpdate struct {
	request.Int64ID
	BrokerCreate
}

type brokerSummary struct {
	ID          int64              `json:"id,string"                gorm:"column:id"`         // broker 节点 ID
	Name        string             `json:"name"                     gorm:"column:name"`       // 名字
	Servername  string             `json:"servername"               gorm:"column:servername"` // servername minion 节点 TLS 认证用
	LAN         []string           `json:"lan"                      gorm:"column:lan;json"`   // 内网地址
	VIP         []string           `json:"vip"                      gorm:"column:vip;json"`   // 外网地址
	Status      bool               `json:"status"                   gorm:"column:status"`     // 状态
	Secret      string             `json:"secret"                   gorm:"column:secret"`     // 随机密钥防止恶意攻击
	Bind        string             `json:"bind"                     gorm:"column:bind"`       // 服务监听地址
	CertID      int64              `json:"cert_id,string,omitempty" gorm:"column:cert_id"`    // 证书 ID
	Semver      string             `json:"semver"                   gorm:"semver"`            // 版本号
	CreatedAt   time.Time          `json:"created_at"               gorm:"column:created_at"` // 创建时间
	UpdatedAt   time.Time          `json:"updated_at"               gorm:"column:updated_at"` // 更新时间
	Certificate *model.Certificate `json:"certificate,omitempty"    gorm:"-"`                 // 证书
}

type BrokerSummaries []*brokerSummary

func (bss BrokerSummaries) CertMap() ([]int64, map[int64]BrokerSummaries) {
	certIDs := make([]int64, 0, 8)
	certMap := make(map[int64]BrokerSummaries, 8)

	for _, sm := range bss {
		certID := sm.CertID
		if certID == 0 {
			continue
		}
		if _, ok := certMap[certID]; !ok {
			certIDs = append(certIDs, certID)
		}
		certMap[certID] = append(certMap[certID], sm)
	}

	return certIDs, certMap
}

type BrokerGoos struct {
	ID      int64  `json:"id,string" gorm:"column:id"`
	Name    string `json:"name"      gorm:"column:name"`
	Linux   int    `json:"linux"     gorm:"column:linux"`
	Windows int    `json:"windows"   gorm:"column:windows"`
	Darwin  int    `json:"darwin"    gorm:"column:darwin"`
}

type BrokerStat struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	MemUsed    uint64  `json:"mem_used"`
	MemTotal   uint64  `json:"mem_total"`
	CPUPercent float64 `json:"cpu_percent"`
}
