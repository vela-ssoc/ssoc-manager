package mrequest

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/vela-ssoc/vela-common-mb/dal/condition"
)

type PageCondition struct {
	Condition
}

type Condition struct {
	CondWhereInputs
	CondOrderInputs
}

func (c Condition) AllInputs() *condition.ScopeInput {
	where := c.CondWhereInputs.Inputs()
	order := c.CondOrderInputs.Inputs()
	return &condition.ScopeInput{
		Where: where,
		Order: order,
	}
}

type CondOrderInputs struct {
	Order []*CondOrderInput `json:"order" form:"order" query:"order" validate:"lte=10,dive"`
}

func (coi CondOrderInputs) Inputs() *condition.OrderInputs {
	orders := coi.Order
	inputs := make([]*condition.OrderInput, 0, len(orders))
	for _, in := range orders {
		if col := in.Name; col != "" {
			data := &condition.OrderInput{Name: col, Desc: in.Desc}
			inputs = append(inputs, data)
		}
	}
	return &condition.OrderInputs{Inputs: inputs}
}

type CondOrderInput struct {
	Name string `json:"name"`
	Desc bool   `json:"desc"`
}

// UnmarshalBind 从 query 参数中解析排序规则。
//
// 支持常规的 JSON 格式：
//   - {"name":"age"}                 ORDER BY age
//   - {"name":"age","desc":false}    ORDER BY age
//   - {"name":"age","desc":true}     ORDER BY age DESC
//
// 支持 <field>:<desc> 格式（因为 JSON 格式放在 query 中太冗长）：
//
//	<desc> 是未定义的值，或者是空（此时冒号可以省略），均按照升序排序：
//		- age:hello    ORDER BY age
//		- age:balabala ORDER BY age
//		- age:         ORDER BY age
//		- age	       ORDER BY age
//	<desc> 可以是枚举：asc、desc：
//		- age:asc      ORDER BY age
//		- age:desc     ORDER BY age DESC
//	<desc> 可以是数字，小于 0 时代表降序，否则代表升序：
//		- age:0        ORDER BY age
//		- age:123      ORDER BY age
//		- age:-1       ORDER BY age DESC
//		- age:-456     ORDER BY age DESC
//
// 错误格式：
//   - :desc
//   - :巴拉巴拉
func (o *CondOrderInput) UnmarshalBind(str string) error {
	if data := []byte(str); json.Valid(data) {
		return json.Unmarshal(data, o)
	}

	idx := strings.LastIndex(str, ":")
	if idx < 0 {
		o.Name = str
		return nil
	} else if idx == 0 {
		return nil
	}

	o.Name = str[:idx]
	value := str[idx+1:]
	if sorted := strings.ToLower(str[idx+1:]); sorted == "asc" || sorted == "desc" {
		o.Desc = sorted == "desc"
		return nil
	}

	num, _ := strconv.ParseInt(value, 10, 64)
	o.Desc = num < 0

	return nil
}

type CondWhereInput struct {
	Name    string   `json:"name"    validate:"required,lte=100"`
	Operate string   `json:"operate" validate:"oneof=eq neq gt gte lt lte like not-like regex not-regex between not-between in not-in"`
	Values  []string `json:"values"  validate:"gt=0,lte=1000"`
}

func (c *CondWhereInput) UnmarshalBind(str string) error {
	if data := []byte(str); json.Valid(data) {
		return json.Unmarshal(data, c)
	}

	// created_at:gte:2024-10-14T06:31:20.620Z
	// size:between:1024,4096
	// hobbies:in:sing,dance,rap

	sn := strings.SplitN(str, ":", 3)
	if len(sn) != 3 {
		return nil
	}
	name, opera, val := sn[0], sn[1], sn[2]
	c.Name, c.Operate = name, opera
	op := condition.NewOperator(opera)
	switch op {
	case condition.Between, condition.NotBetween,
		condition.In, condition.NotIn:
		c.Values = strings.Split(val, ",")
	default:
		c.Values = []string{val}
	}

	return nil
}

type CondWhereInputs struct {
	Or    bool              `json:"or"    form:"or"    query:"or"`
	Where []*CondWhereInput `json:"where" form:"where" query:"where" validate:"lte=100,dive"`
}

func (cws CondWhereInputs) Inputs() *condition.WhereInputs {
	size := len(cws.Where)
	inputs := make([]*condition.WhereInput, 0, size)
	for _, cw := range cws.Where {
		if col := cw.Name; col != "" {
			data := &condition.WhereInput{
				Name:    col,
				Operate: condition.NewOperator(cw.Operate),
				Values:  cw.Values,
			}
			inputs = append(inputs, data)
		}
	}
	return &condition.WhereInputs{
		Or:     cws.Or,
		Inputs: inputs,
	}
}
