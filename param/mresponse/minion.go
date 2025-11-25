package mresponse

import (
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
)

type MinionItem struct {
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
	IBU        string             `json:"ibu"         gorm:"column:ibu"`
	Comment    string             `json:"comment"     gorm:"column:comment"`
	BrokerName string             `json:"broker_name" gorm:"column:broker_name"`
	Unload     bool               `json:"unload"      gorm:"column:unload"`
	Uptime     time.Time          `json:"uptime"      gorm:"column:uptime"`
	Unstable   bool               `json:"unstable"    gorm:"column:unstable"`
	Customized string             `json:"customized"  gorm:"column:customized"`
	OSRelease  string             `json:"os_release"  gorm:"column:os_release"`
	Tags       []string           `json:"tags"        gorm:"-"`
}
