package mresponse

type EffectProgress struct {
	ID       int64 `json:"id,string"` // 任务 ID
	Count    int   `json:"count"`     // 总数
	Executed int   `json:"executed"`  // 已经下发完毕的
	Failed   int   `json:"failed"`    // 下发失败的
}
