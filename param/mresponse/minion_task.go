package mresponse

import (
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
)

// MinionTask 节点关联的配置。
//
// - 如果节点是静默模式，则所有的 report 应该为 null，不为 null 代表配置不同步。
// - 如果配置的 excluded 为 true，则对应的 report 应该为 null，不为 null 代表配置不同步。
// - 其他情况下，tasks.$.hash 必须与 tasks.$.report.hash 相等才算配置同步。
type MinionTask struct {
	Unload bool              `json:"unload"` // 是否静默模式
	Tasks  []*MinionTaskItem `json:"tasks"`  // 关联的配置
}

type MinionTaskItem struct {
	ID           int64                 `json:"id,string"`      // ID
	Name         string                `json:"name"`           // 名字
	Icon         []byte                `json:"icon"`           // 图标
	Dialect      bool                  `json:"dialect"`        // 是否私有模块。
	Excluded     bool                  `json:"excluded"`       // 是否被排除（一对一）。
	ExcludedInet bool                  `json:"excluded_inet"`  // 是否被排除（旧版设计方案）。
	Hash         string                `json:"hash"`           // 哈希（md5）
	Desc         string                `json:"desc"`           // 描述
	Report       *MinionTaskItemReport `json:"report"`         // agent 上报信息
	Chunk        []byte                `json:"chunk,omitzero"` // 代码
	Version      int64                 `json:"version"`        // 代码版本
	CreatedAt    time.Time             `json:"created_at"`     // 创建时间
	UpdatedAt    time.Time             `json:"updated_at"`     // 修改时间
}

type MinionTaskItemReport struct {
	From      string            `json:"from"`       // tunnel 为中心端下发，其他为 agent 自定义。
	Uptime    time.Time         `json:"uptime"`     // 启动时间
	Link      string            `json:"link"`       // 外链（程序依赖另一个程序）
	Status    string            `json:"status"`     // 运行状态
	Hash      string            `json:"hash"`       // 哈希（md5）
	Cause     string            `json:"cause"`      // 错误原因（如果有）
	Runners   model.TaskRunners `json:"runners"`    // 运行时内部依赖
	CreatedAt time.Time         `json:"created_at"` // 上报时间
}
