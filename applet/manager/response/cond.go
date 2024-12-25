package response

import "github.com/vela-ssoc/vela-common-mb/dal/condition"

type Cond struct {
	Wheres []*CondWhere `json:"wheres,omitempty"`
	Orders []*CondOrder `json:"orders,omitempty"`
}

type CondOrder struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

type CondWhere struct {
	Name      string   `json:"name"`
	Comment   string   `json:"comment"`
	Operators []string `json:"operators"`
}

func ReadCond(c *condition.Cond) *Cond {
	ret := new(Cond)
	if c == nil {
		return ret
	}

	for _, f := range c.OrderFields() {
		name, comment := f.NameComment()
		data := &CondOrder{Name: name, Comment: comment}
		ret.Orders = append(ret.Orders, data)
	}
	for _, f := range c.WhereFields() {
		name, comment := f.NameComment()
		data := &CondWhere{Name: name, Comment: comment}
		ops := f.Operators()
		for _, op := range ops {
			data.Operators = append(data.Operators, op.String())
		}

		ret.Wheres = append(ret.Wheres, data)
	}

	return ret
}
