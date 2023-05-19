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
