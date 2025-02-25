package param

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/vela-ssoc/vela-common-mb/param/request"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
)

type TaskExtensionCreate struct {
	Name        string          `json:"name"                validate:"required,lte=100"`
	Intro       string          `json:"intro"               validate:"required,lte=255"`
	Content     string          `json:"content"             validate:"lte=65535"`
	ExtensionID int64           `json:"extension_id,string" validate:"required_without=Content"`
	Data        json.RawMessage `json:"data"` // 引用插件商店时的
}

type TaskExtensionRelease struct {
	ID            int64          `json:"id,string"      validate:"required"`
	Name          string         `json:"name,string"    validate:"required,lte=100"`
	Intro         string         `json:"intro,string"   validate:"required,lte=1000"`
	WindowSize    int            `json:"window_size"    validate:"gte=0,lte=10000"`
	Timeout       model.Duration `json:"timeout"`
	Cron          string         `json:"cron"           validate:"cron"`
	SpecificTimes []time.Time    `json:"specific_times" validate:"lte=100"`
	Filters       []string       `json:"filters"        validate:"lte=100,dive,required,lte=100"`
	Excludes      []string       `json:"excludes"       validate:"lte=100,dive,required,lte=100"`
}

type TaskExtensionFromMarket struct {
	Name        string          `json:"name"                validate:"required,lte=100"`
	Intro       string          `json:"intro"               validate:"required,lte=255"`
	ExtensionID int64           `json:"extension_id,string" validate:"required_without=Content"`
	Data        json.RawMessage `json:"data"` // 引用插件商店时的数据
}

type TaskExtensionFromCode struct {
	Name  string `json:"name"  validate:"required,lte=100"`
	Intro string `json:"intro" validate:"required,lte=255"`
	Code  string `json:"code"  validate:"required,lte=65535"`
}

type TaskExtensionCreateCode struct {
	Name        string          `json:"name"                validate:"required,lte=100"`
	Intro       string          `json:"intro"               validate:"lte=1000"`
	Code        string          `json:"code"                validate:"lte=65535"`
	ExtensionID int64           `json:"extension_id,string" validate:"required_without=Code"`
	Data        json.RawMessage `json:"data"`
}

type TaskExtensionUpdateCode struct {
	IntID
	Intro       string          `json:"intro"               validate:"lte=1000"`
	Code        string          `json:"code"                validate:"lte=65535"`
	ExtensionID int64           `json:"extension_id,string" validate:"required_without=Code"`
	Data        json.RawMessage `json:"data"`
}

type TaskExtensionCreatePublish struct {
	Name          string          `json:"name"                validate:"required,lte=100"`
	Intro         string          `json:"intro"               validate:"required,lte=1000"`
	Code          string          `json:"code"                validate:"lte=65535"`
	ExtensionID   int64           `json:"extension_id,string" validate:"required_without=Code"`
	Data          json.RawMessage `json:"data"`
	PushSize      int             `json:"push_size"           validate:"gte=1,lte=10000"`
	Timeout       model.Duration  `json:"timeout"`
	Cron          string          `json:"cron"                validate:"omitempty,cron"`
	SpecificTimes []time.Time     `json:"specific_times"      validate:"lte=100"`
	Enabled       bool            `json:"enabled"`
	Filters       []string        `json:"filters"             validate:"lte=100,dive,required,lte=100"`
	Excludes      []string        `json:"excludes"            validate:"lte=100,dive,required,lte=100"`
}

type TaskExtensionUpdatePublish struct {
	IntID
	Intro         string          `json:"intro"               validate:"required,lte=1000"`
	Code          string          `json:"code"                validate:"lte=65535"`
	ExtensionID   int64           `json:"extension_id,string" validate:"required_without=Code"`
	Data          json.RawMessage `json:"data"`
	PushSize      int             `json:"push_size"           validate:"gte=1,lte=10000"`
	Timeout       model.Duration  `json:"timeout"`
	Cron          string          `json:"cron"                validate:"omitempty,cron"`
	SpecificTimes Times           `json:"specific_times"      validate:"lte=100"`
	Enabled       bool            `json:"enabled"`
	Filters       Strings         `json:"filters"             validate:"lte=100,dive,required,lte=100"`
	Excludes      Strings         `json:"excludes"            validate:"lte=100,dive,required,lte=100"`
}

type TaskExtensionPublishFilter struct {
	request.Keywords
	request.CondWhereInputs
	Tags    []string `json:"tags"    validate:"lte=1000"`
	TagMode bool     `json:"tag_mode"`
}

type Strings []string

func (s Strings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Strings) Scan(src any) error {
	bs, _ := src.([]byte)
	return json.Unmarshal(bs, s)
}

type Times []time.Time

func (ts *Times) Scan(src any) error {
	bs, _ := src.([]byte)
	return json.Unmarshal(bs, ts)
}

func (ts Times) Value() (driver.Value, error) {
	return json.Marshal(ts)
}
