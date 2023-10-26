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

func (m MinionTaskSummary) String() string {
	return m.Name + "[" + m.Status + "]"
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

type TaskList []*taskList

type taskList struct {
	ID          int64     `json:"id,string"           gorm:"column:id"`           // 数据库 ID
	Inet        string    `json:"inet"                gorm:"column:inet"`         // 节点 IPv4
	MinionID    int64     `json:"minion_id,string"    gorm:"column:minion_id"`    // 节点 ID
	SubstanceID int64     `json:"substance_id,string" gorm:"column:substance_id"` // 配置 ID
	Name        string    `json:"name"                gorm:"column:name"`         // 配置名
	Dialect     bool      `json:"dialect"             gorm:"column:dialect"`      // 是否时私有配置
	Status      string    `json:"status"              gorm:"column:status"`       // 运行状态 running doing fail panic reg update
	Hash        string    `json:"hash"                gorm:"column:hash"`         // 配置文件哈希
	ReportHash  string    `json:"report_hash"         gorm:"column:report_hash"`  // 上报的哈希
	Link        string    `json:"link"                gorm:"column:link"`         // 外链
	From        string    `json:"from"                gorm:"column:from"`         // 来源模块
	Failed      bool      `json:"failed"              gorm:"column:failed"`       // 是否运行错误
	Cause       string    `json:"cause"               gorm:"column:cause"`        // 运行错误的原因
	UpdatedAt   time.Time `json:"updated_at"          gorm:"column:updated_at"`   // 配置文件最近更新事件
	CreatedAt   time.Time `json:"created_at"          gorm:"column:created_at"`   // 配置文件创建事件
	ReportAt    time.Time `json:"report_at"           gorm:"column:report_at"`    // 下发节点的最新上报时间
}

type TaskRCount struct {
	ID    int64  `json:"id,string" gorm:"column:id"`
	Name  string `json:"name"      gorm:"column:name"`
	Desc  string `json:"desc"      gorm:"-"`
	Count int    `json:"count"     gorm:"column:count"`
}
