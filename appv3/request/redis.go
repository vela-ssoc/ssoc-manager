package request

type RedisKeys struct {
	Size
	Match  string `json:"match"  query:"match"`
	Cursor uint64 `json:"cursor" query:"cursor"`
}

type RedisValue struct {
	Size
	Key    string `json:"key"    query:"key"    validate:"required"`
	Match  string `json:"match"  query:"match"`
	Cursor uint64 `json:"cursor" query:"cursor"`
}
