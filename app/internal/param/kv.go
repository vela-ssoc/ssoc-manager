package param

import "gorm.io/gen/field"

type IDName struct {
	ID   int64  `json:"id,string" gorm:"column:id"`
	Name string `json:"name"      gorm:"column:name"`
}

type IDNames []*IDName

func (dn IDNames) Map() map[int64]*IDName {
	hm := make(map[int64]*IDName, len(dn))
	for _, n := range dn {
		hm[n.ID] = n
	}
	return hm
}

type NameCount struct {
	Name  string `json:"name"  gorm:"column:name"`
	Count int    `json:"count" gorm:"column:count"`
}

type NameCounts []*NameCount

func (nc NameCounts) Aliases() (name, count field.Field) {
	// 要与 NameCount gorm tag 保持一致
	name = field.NewField("", "name")
	count = field.NewField("", "count")
	return
}
