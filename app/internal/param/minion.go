package param

import (
	"time"

	"github.com/vela-ssoc/vela-common-mb/dynsql"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
)

type MinionCreate struct {
	Inet string `json:"inet" validate:"ipv4"`
	Goos string `json:"goos" validate:"oneof=linux windows darwin"`
	Arch string `json:"arch" validate:"oneof=amd64 386 arm64 arm"`
}

type MinionSummary struct {
	ID         int64              `json:"id,string"`
	Inet       string             `json:"inet"`
	Goos       string             `json:"goos"`
	Edition    string             `json:"edition"`
	Status     model.MinionStatus `json:"status"`
	CPUCore    int                `json:"cpu_core"`
	MemTotal   int                `json:"mem_total"`
	MemFree    int                `json:"mem_free"`
	DiskTotal  int                `json:"disk_total"`
	DiskFree   int                `json:"disk_free"`
	IDC        string             `json:"idc"`
	IBu        string             `json:"ibu"`
	Comment    string             `json:"comment"`
	BrokerName string             `json:"broker_name"`
	Unload     bool               `json:"unload"` // 是否不加载配置模式
	Uptime     time.Time          `json:"uptime"` // 最近一次上线时间
	Tags       []string           `json:"tags"`
}

type MinionDetail struct {
	ID            string             `json:"id"             gorm:"column:id"`
	Inet          string             `json:"inet"           gorm:"column:inet"`
	Inet6         string             `json:"inet6"          gorm:"column:inet6"`
	MAC           string             `json:"mac"            gorm:"column:mac"`
	Goos          string             `json:"goos"           gorm:"column:goos"`
	Arch          string             `json:"arch"           gorm:"column:arch"`
	Edition       string             `json:"edition"        gorm:"column:edition"`
	Status        model.MinionStatus `json:"status"         gorm:"column:status"`
	Uptime        time.Time          `json:"uptime"         gorm:"column:minion.uptime"`
	BrokerID      string             `json:"broker_id"      gorm:"column:broker_id"`
	BrokerName    string             `json:"broker_name"    gorm:"column:broker_name"`
	OrgPath       string             `json:"org_path"       gorm:"column:org_path"`
	Identity      string             `json:"identity"       gorm:"column:identity"`
	Category      string             `json:"category"       gorm:"column:category"`
	OpDuty        string             `json:"op_duty"        gorm:"column:op_duty"`
	Comment       string             `json:"comment"        gorm:"column:comment"`
	IBu           string             `json:"ibu"            gorm:"column:ibu"`
	IDC           string             `json:"idc"            gorm:"column:idc"`
	CreatedAt     time.Time          `json:"created_at"     gorm:"column:created_at"`
	UpdatedAt     time.Time          `json:"updated_at"     gorm:"column:updated_at"`
	Release       string             `json:"release"        gorm:"column:release"`
	CPUCore       int                `json:"cpu_core"       gorm:"column:cpu_core"`
	MemTotal      int64              `json:"mem_total"      gorm:"column:mem_total"`
	MemFree       int64              `json:"mem_free"       gorm:"column:mem_free"`
	SwapTotal     int64              `json:"swap_total"     gorm:"column:swap_total"`
	SwapFree      int64              `json:"swap_free"      gorm:"column:swap_free"`
	DiskPath      int64              `json:"disk_path"      gorm:"column:disk_path"`
	DiskTotal     int64              `json:"disk_total"     gorm:"column:disk_total"`
	DiskFree      int64              `json:"disk_free"      gorm:"column:disk_free"`
	HostID        string             `json:"host_id"        gorm:"column:host_id"`
	Family        string             `json:"family"         gorm:"column:family"`
	Version       string             `json:"version"        gorm:"column:version"`
	Elapse        string             `json:"elapse"         gorm:"column:elapse"`
	BootAt        string             `json:"boot_at"        gorm:"column:boot_at"`
	VirtualSys    string             `json:"virtual_sys"    gorm:"column:virtual_sys"`
	VirtualRole   string             `json:"virtual_role"   gorm:"column:virtual_role"`
	ProcNumber    int                `json:"proc_number"    gorm:"column:proc_number"`
	Hostname      string             `json:"hostname"       gorm:"column:hostname"`
	KernelVersion string             `json:"kernel_version" gorm:"column:kernel_version"`
	AgentTotal    int64              `json:"agent_total"    gorm:"column:agent_total"`
	AgentAlloc    int64              `json:"agent_alloc"    gorm:"column:agent_alloc"`
	Unload        bool               `json:"unload"         gorm:"column:unload"`
	Tags          model.MinionTags   `json:"tags"           gorm:"-"`
}

type MinionBatchRequest struct {
	dynsql.Input
	Cmd     string `json:"cmd"     query:"cmd"     validate:"oneof=resync restart upgrade offline"`
	Keyword string `json:"keyword" query:"keyword"`
}

type MinionCommandRequest struct {
	IntID
	Command string `json:"command" validate:"oneof=resync restart upgrade offline"`
}

type MinionUnloadRequest struct {
	IntID
	Unload bool `json:"unload"`
}

type MinionDeleteRequest struct {
	dynsql.Input
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}

func (k MinionDeleteRequest) Like() string {
	if k.Keyword == "" {
		return ""
	}
	return "%" + k.Keyword + "%"
}

type MinionUpgradeRequest struct {
	IntID
	Semver model.Semver `json:"semver" validate:"omitempty,semver"`
}
