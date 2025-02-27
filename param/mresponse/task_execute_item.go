package mresponse

type TaskExecuteItemCodeCount struct {
	Code  int   `json:"code"  gorm:"column:code"`
	Count int64 `json:"count" gorm:"column:count"`
}

type TaskExecuteItemCodeCounts []*TaskExecuteItemCodeCount

func (ts TaskExecuteItemCodeCounts) Aliases() (name, count string) {
	// 要与 TaskExecuteItemCodeCount gorm tag 保持一致
	return "code", "count"
}
