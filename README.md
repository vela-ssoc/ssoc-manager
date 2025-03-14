# [ssoc](https://github.com/vela-ssoc)

## 贯标

### 常用字段数据库定义

GOOS: varchar(10)

GOARCH: VARCHAR(15)

## minion 版本管理和升级规则

minion 表结构体定义如下：

```go
package model

import (
	"strconv"
	"strings"
	"time"
)

// MinionBin minion 节点二进制发行版本记录表
type MinionBin struct {
	ID         int64     `json:"id,string"  gorm:"column:id;primaryKey"` // 表 ID
	FileID     int64     `json:"-"          gorm:"column:file_id"`       // 关联文件表 ID
	Goos       string    `json:"goos"       gorm:"column:goos"`          // 操作系统
	Arch       string    `json:"arch"       gorm:"column:arch"`          // 系统架构
	Name       string    `json:"name"       gorm:"column:name"`          // 文件名称
	Customized string    `json:"customized" gorm:"column:customized"`    // 定制版标记
	Unstable   bool      `json:"unstable"   gorm:"column:unstable"`      // 不稳定版本，内测版本
	Caution    string    `json:"caution"    gorm:"column:caution"`       // 注意事项
	Ability    string    `json:"ability"    gorm:"column:ability"`       // 功能说明
	Size       int64     `json:"size"       gorm:"column:size"`          // 文件大小
	Hash       string    `json:"hash"       gorm:"column:hash"`          // 文件哈希
	Semver     Semver    `json:"semver"     gorm:"column:semver"`        // 版本号
	Changelog  string    `json:"changelog"  gorm:"column:changelog"`     // 更新日志
	Weight     int64     `json:"-"          gorm:"column:weight"`        // 版本号权重，用于比较版本号大小
	Deprecated bool      `json:"deprecated" gorm:"column:deprecated"`    // 是否已过期
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName implement gorm schema.Tabler
func (MinionBin) TableName() string {
	return "minion_bin"
}

// Semver https://semver.org/lang/zh-CN/
type Semver string

// Int64 计算版本号
func (sv Semver) Int64() int64 {
	sn := strings.SplitN(string(sv), ".", 3)
	if len(sn) != 3 {
		return 0
	}

	sp := strings.SplitN(sn[2], "-", 2)
	sn[2] = sp[0]

	var ret int64
	for _, s := range sn {
		num, _ := strconv.ParseInt(s, 10, 64)
		ret *= 1000000
		ret += num
	}

	return ret
}
```

上面的结构体定义截取自 [Git 版本](https://github.com/vela-ssoc/ssoc-common-mb/blob/4181ff7335550131dc7cff52a22652bb2469d17e/dal/model/minion_bin.go)

> 随着时间发展，字段或业务可能会有调整，对于版本管理和节点升级逻辑，该文档依然有参考价值。

该结构体中比较重要的字段是：

- `Goos`: 操作系统类型 `runtime.GOOS`

- `Arch`: 操作系统架构 `runtime.GOARCH`

- `Semver`: 版本号，如：1.2.3、1.2.3-beta

- `Weigth`: 由 `Semver` 计算而来，用于版本号比较和排序。注：数据库中的直接拿 `Semver` 排序是字典序，`1.1.9` 比 `1.1.10`
  靠前，故将版本号转化为数字方便版本号之间比较。

- `Customized`: 定制标签，如：扫描器，抓包器等。该字段为空代表是标准版。通常情况下定制版对环境依赖多于标准版，比如需要特定的操作系统内核版本、需要操作系统预先安装一些
  lib 库等等。所以标准版的通用性要好一些。定制版本可以执行批量升级，但是升级范围只波及同样 `Goos` + `Arch` + `Customized`
  的版本，注：标准版也同理，标准版批量升级操作也只会波及标准版节点。

- `Unstable`: 是否是测试版，该字段在上传二进制时已经被确定，且不可修改。测试版不可以批量升级，只允许显式指定节点推送升级，任何节点都可以手动推送升级到测试版。

- `Deprecated`: 强烈反对的版本，通常是在发行后发现较大的 BUG 或其它问题的版本，一旦被标记为 `Deprecated` 后不可反向标记回来。
  该版本在任何情况下都不会下发给 minion 节点升级。

- `Goos` + `Arch` + `Semver` + `Customized` 四者组合唯一（联合主键）。

- 如果上传的是 `Unstable` 版本，请上传者在命名版本号时添加 __先行版本号__
  ，尽量避免占用正式版本号，比如可以叫作：`1.0.0-alpha`、`1.0.0-beta` 等，把 `1.0.0`
  留给正式版命名。此处可参考 [语义化版本控制规范 第九条](https://semver.org/lang/zh-CN/#spec-item-9)。

- 由于已经有了 `Unstable` 字段作为区分测试版的标记，后台程序就不再根据版本号中是否有 __先行版本号__
  来区分是否是测试版（相对于上次的设计）。即：现在的版本完全根据 `Unstable` 来判断是否是测试版。

- 单节点手动升级可以跨 `Unstable`、`Customized`，请操作者升级时考虑好存在风险。比如：标准版与定制版之间可以互相升级；测试版和稳定版之间可以互相升级。

- 对于全局性的批量升级（节点分页列表右上角的批量升级按钮），由于节点只会收到升级信号，报文中不包含其他信息。此时中心会根据当前节点的正在运行的版本情况进行推送：1. `Unstable` 节点不会升级。 2. 稳定版本按照 `Goos` + `Arch` + `Customized` 匹配找到最新版本下发升级。

- 节点升级的大前提是：升级包的 `Goos` `Arch` 与 minion 节点当前的 `Goos` `Arch` 一致。上述虽未提及该规则，但要牢记悉知。