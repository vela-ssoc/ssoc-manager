package param

import (
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
)

type MinionCreate struct {
	Inet string `json:"inet" validate:"ipv4"`
	Goos string `json:"goos" validate:"oneof=linux windows darwin"`
	Arch string `json:"arch" validate:"oneof=amd64 386 arm64 arm loong64 riscv64"`
}

type MinionSummary struct {
	ID         int64              `json:"id,string"   gorm:"column:id"`
	Inet       string             `json:"inet"        gorm:"column:inet"`
	Goos       string             `json:"goos"        gorm:"column:goos"`
	Edition    string             `json:"edition"     gorm:"column:edition"`
	Status     model.MinionStatus `json:"status"      gorm:"column:status"`
	CPUCore    int                `json:"cpu_core"    gorm:"-"`
	MemTotal   int                `json:"mem_total"   gorm:"-"`
	MemFree    int                `json:"mem_free"    gorm:"-"`
	DiskTotal  int                `json:"disk_total"  gorm:"-"`
	DiskFree   int                `json:"disk_free"   gorm:"-"`
	IDC        string             `json:"idc"         gorm:"column:idc"`
	IBu        string             `json:"ibu"         gorm:"column:ibu"`
	Comment    string             `json:"comment"     gorm:"column:comment"`
	BrokerName string             `json:"broker_name" gorm:"column:broker_name"`
	Unload     bool               `json:"unload"      gorm:"column:unload"`
	Uptime     time.Time          `json:"uptime"      gorm:"column:uptime"`
	Unstable   bool               `json:"unstable"    gorm:"column:unstable"`
	Customized string             `json:"customized"  gorm:"column:customized"`
	Tags       []string           `json:"tags"        gorm:"-"`
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
	Uptime        time.Time          `json:"uptime"         gorm:"column:uptime"`
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
	MemTotal      int                `json:"mem_total"      gorm:"column:mem_total"`
	MemFree       int                `json:"mem_free"       gorm:"column:mem_free"`
	SwapTotal     int                `json:"swap_total"     gorm:"column:swap_total"`
	SwapFree      int                `json:"swap_free"      gorm:"column:swap_free"`
	HostID        string             `json:"host_id"        gorm:"column:host_id"`
	Family        string             `json:"family"         gorm:"column:family"`
	Version       string             `json:"version"        gorm:"column:version"`
	BootAt        int64              `json:"boot_at"        gorm:"column:boot_at"`
	VirtualSys    string             `json:"virtual_sys"    gorm:"column:virtual_sys"`
	VirtualRole   string             `json:"virtual_role"   gorm:"column:virtual_role"`
	ProcNumber    int                `json:"proc_number"    gorm:"column:proc_number"`
	Hostname      string             `json:"hostname"       gorm:"column:hostname"`
	KernelVersion string             `json:"kernel_version" gorm:"column:kernel_version"`
	AgentTotal    int                `json:"agent_total"    gorm:"column:agent_total"`
	AgentAlloc    int                `json:"agent_alloc"    gorm:"column:agent_alloc"`
	Unload        bool               `json:"unload"         gorm:"column:unload"`
	Unstable      bool               `json:"unstable"       gorm:"column:unstable"`   // 是否不稳定版本
	Customized    string             `json:"customized"     gorm:"column:customized"` // 定制版
	Tags          model.MinionTags   `json:"tags"           gorm:"-"`
}

type MinionBatchRequest struct {
	dynsql.Input
	Cmd     string `json:"cmd"     query:"cmd"     validate:"oneof=resync restart upgrade offline"`
	Keyword string `json:"keyword" query:"keyword"`
}

func (r MinionBatchRequest) Like() string {
	key := r.Keyword
	if key == "" {
		return ""
	}
	return "%" + key + "%"
}

type MinionUnloadRequest struct {
	request.Int64ID
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
	ID       int64 `json:"id,string"        validate:"required"`
	BinaryID int64 `json:"binary_id,string" validate:"required"`
}

type MinionTagRequest struct {
	dynsql.Input
	Keyword string   `json:"keyword" query:"keyword" form:"keyword"`
	Deletes []string `json:"deletes" validate:"lte=10,dive,tag"`
	Creates []string `json:"creates" validate:"lte=10,dive,tag"`
}

func (k MinionTagRequest) Like() string {
	if k.Keyword == "" {
		return ""
	}
	return "%" + k.Keyword + "%"
}
