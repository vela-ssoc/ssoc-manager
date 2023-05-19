package param

type IDName struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}

type NameCount struct {
	Name  string `json:"name"  group:"column:name"`
	Count int    `json:"count" group:"column:count"`
}
