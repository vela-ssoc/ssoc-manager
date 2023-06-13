package param

type NameCount struct {
	Name  string `json:"name"  group:"column:name"`
	Count int    `json:"count" group:"column:count"`
}

type IDName struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}

type IDNames []*IDName

func (dn IDNames) Map() map[int64]*IDName {
	hm := make(map[int64]*IDName, len(dn))
	for _, n := range dn {
		hm[n.ID] = n
	}
	return hm
}
