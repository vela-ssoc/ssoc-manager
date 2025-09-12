package param

import (
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
)

type SubstanceSummary struct {
	ID        int64     `json:"id,string"`
	Name      string    `json:"name"`
	Icon      []byte    `json:"icon"`
	Hash      string    `json:"hash"`
	Desc      string    `json:"desc"`
	Links     []string  `json:"links"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubstanceCreate struct {
	Name     string `json:"name"  validate:"substance_name"` // startup 是每个minion的启动配置名字
	Desc     string `json:"desc"  validate:"lte=200"`
	Icon     []byte `json:"icon"  validate:"lte=65536"`
	Chunk    []byte `json:"chunk" validate:"gt=0,lte=524288"` // 524288 = 512 * 1024, 512k
	MinionID int64  `json:"minion_id,string"`
	Priority int64  `json:"priority"` // 优先级，越大越先执行
}

type SubstanceUpdate struct {
	ID       int64  `json:"id,string"`
	Desc     string `json:"desc"  validate:"lte=200"`
	Icon     []byte `json:"icon"  validate:"omitempty,lte=65536"`
	Chunk    []byte `json:"chunk" validate:"gt=0,lte=524288"` // 524288 = 512 * 1024, 512k
	Version  int64  `json:"version"`
	Priority int64  `json:"priority"` // 优先级，越大越先执行
}

type SubstanceReload struct {
	request.Int64ID
	SubstanceID int64 `json:"substance_id,string" validate:"required,gt=0"`
}

type SubstanceCommand struct {
	request.Int64ID
	Cmd string `json:"cmd" validate:"oneof=resync offline"`
}

type IDPageSQL struct {
	OptionalID
	PageSQL
}
