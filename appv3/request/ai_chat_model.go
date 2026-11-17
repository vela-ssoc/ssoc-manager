package request

type AIChatModelUpsert struct {
	AIChatModelModels
	Enabled bool     `json:"enabled"`                                           // 是否启用
	Models  []string `json:"models"   validate:"lte=1000,unique,dive,required"` // 允许使用的模型
	Prompt  string   `json:"prompt"   validate:"lte=10000"`                     // 默认系统提示词
}

type AIChatModelModels struct {
	BaseURL string `json:"base_url" validate:"required,http_url"` // BaseURL
	APIKey  string `json:"api_key"  validate:"required"`          // APIKey
}
