package param

type BlinkAlert struct {
	JobNumbers []string `json:"job_numbers"` // 接收人工号
	RuleID     string   `json:"rule_id"`     // 规则 ID
	Title      string   `json:"title"`       // 广播消息存在标题
	Detail     string   `json:"detail"`      // 广播消息体或卡片消息内容
}
