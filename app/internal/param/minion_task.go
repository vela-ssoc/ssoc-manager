package param

import (
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
)

type MinionTaskSummary struct {
	ID         int64     `json:"id,string"`   // ID
	Name       string    `json:"name"`        // 名字
	Icon       []byte    `json:"icon"`        // 图标
	From       string    `json:"from"`        // 来源模块
	Status     string    `json:"status"`      // 运行状态
	Link       string    `json:"link"`        // 外链
	Dialect    bool      `json:"dialect"`     // 是否私有配置
	LegalHash  string    `json:"legal_hash"`  // 数据库记录的 hash
	ActualHash string    `json:"actual_hash"` // 节点上上报的 hash
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
	UpdatedAt  time.Time `json:"updated_at"`  // 修改时间
}

type MinionTaskDetail struct {
	ID         int64             `json:"id,string"`   // ID
	Name       string            `json:"name"`        // 名字
	Icon       []byte            `json:"icon"`        // 图标
	From       string            `json:"from"`        // 来源模块
	Status     string            `json:"status"`      // 运行状态
	Link       string            `json:"link"`        // 外链
	Desc       string            `json:"desc"`        // 描述
	Dialect    bool              `json:"dialect"`     // 是否私有配置
	LegalHash  string            `json:"legal_hash"`  // 数据库记录的 hash
	ActualHash string            `json:"actual_hash"` // 节点上上报的 hash
	Failed     bool              `json:"failed"`      // 是否运行失败
	Cause      string            `json:"cause"`       // 如果失败，失败原因
	Chunk      []byte            `json:"chunk"`
	Version    int64             `json:"version"`
	CreatedAt  time.Time         `json:"created_at"` // 创建时间
	UpdatedAt  time.Time         `json:"updated_at"` // 修改时间
	TaskAt     time.Time         `json:"task_at"`    // 上报时间
	Uptime     time.Time         `json:"uptime"`     // 启动时间
	Runners    model.TaskRunners `json:"runners"`    // 内部服务
}

type MinionTaskDetailRequest struct {
	IntID
	SubstanceID int64 `json:"substance_id,string" query:"substance_id"`
}

type TaskGather struct {
	Name    string `json:"name"`
	Dialect bool   `json:"dialect"`
	Running int    `json:"running"`
	Doing   int    `json:"doing"`
	Fail    int    `json:"fail"`
	Panic   int    `json:"panic"`
	Reg     int    `json:"reg"`
	Update  int    `json:"update"`
}

// TaskCount task 状态分类统计
type TaskCount struct {
	Dialect int `json:"dialect" gorm:"column:dialect"` // 私有配置总数
	Public  int `json:"public"  gorm:"column:public"`  // 公有配置总数
	Running int `json:"running" gorm:"column:running"` // 运行状态为 running 的总数
	Doing   int `json:"doing"   gorm:"column:doing"`   // 运行状态为 doing   的总数
	Fail    int `json:"fail"    gorm:"column:fail"`    // 运行状态为 fail    的总数
	Panic   int `json:"panic"   gorm:"column:panic"`   // 运行状态为 panic   的总数
	Reg     int `json:"reg"     gorm:"column:reg"`     // 运行状态为 reg     的总数
	Update  int `json:"update"  gorm:"column:update"`  // 运行状态为 update  的总数
}
