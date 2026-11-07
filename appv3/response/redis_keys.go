package response

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisKeys struct {
	Records []RedisKeyMeta `json:"keys"`
	Cursor  uint64         `json:"cursor"`
}

type RedisKeyMeta struct {
	Key    string        `json:"key"`
	Type   string        `json:"type"`
	TTL    time.Duration `json:"ttl"`
	Memory int64         `json:"memory"`
}

type RedisKeyStatusCmd struct {
	Type   *redis.StatusCmd
	TTL    *redis.DurationCmd
	Memory *redis.IntCmd
}

type RedisValue struct {
	Key    string `json:"key"`
	Type   string `json:"type"`
	Data   any    `json:"data"`
	Cursor uint64 `json:"cursor"`
}
