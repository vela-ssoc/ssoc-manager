package service

import (
	"context"
	"log/slog"
	"maps"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/memcache"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type RedisConfig struct {
	db    polydb.Sourcer
	log   *slog.Logger
	mem   *memcache.Value[redis.UniversalClient]
	safes map[string]struct{}
}

func NewRedisConfig(db polydb.Sourcer, log *slog.Logger) *RedisConfig {
	rc := &RedisConfig{db: db, log: log}
	rc.mem = memcache.NewValue(rc.slowLoad)
	rc.safes = rc.safeCommands()

	return rc
}

func (rc *RedisConfig) Get(ctx context.Context) (*model.RedisConfig, error) {
	mdb := rc.db.MDB()
	coll := mdb.RedisConfig()

	return coll.Get(ctx)
}

func (rc *RedisConfig) Upsert(ctx context.Context, param request.RedisConfigUpsert) error {
	now := time.Now()
	mod := &model.RedisConfig{
		Enabled:     param.Enabled,
		Addresses:   param.Addresses,
		Username:    param.Username,
		Password:    param.Password,
		ClusterMode: param.ClusterMode,
		MasterName:  param.MasterName,
		DB:          param.DB,
		UpdatedAt:   now,
	}
	mdb := rc.db.MDB()
	coll := mdb.RedisConfig()
	opts := options.UpdateOne().SetUpsert(true)
	update := bson.M{"$set": mod, "$setOnInsert": bson.M{"created_at": now}}
	_, err := coll.UpdateOne(ctx, bson.D{}, update, opts)
	if err == nil {
		rc.Forget()
	}

	return err
}

func (rc *RedisConfig) Info(ctx context.Context, sections []string) (string, error) {
	cli, err := rc.mem.Load(ctx)
	if err != nil {
		return "", err
	}

	return cli.Info(ctx, sections...).Result()
}

func (rc *RedisConfig) Do(ctx context.Context, args []any) (any, error) {
	cli, err := rc.mem.Load(ctx)
	if err != nil {
		return "", err
	}

	val, err := cli.Do(ctx, args...).Result()
	if err == nil {
		rc.log.Info("执行 redis 命令成功", "args", args)
	} else {
		rc.log.Info("执行 redis 命令出错", "args", args, "err", err)
	}

	return val, err
}

func (rc *RedisConfig) Keys(ctx context.Context, cursor uint64, match string, count int64) (*response.RedisKeys, error) {
	cli, err := rc.mem.Load(ctx)
	if err != nil {
		return nil, err
	}

	keys, next, err := cli.Scan(ctx, cursor, match, count).Result()
	if err != nil {
		return nil, err
	}
	ret := &response.RedisKeys{Cursor: next}
	if len(keys) == 0 {
		return ret, nil
	}

	tmps := make([]*response.RedisKeyStatusCmd, len(keys))
	pipe := cli.Pipeline()
	for i, key := range keys {
		tmp := &response.RedisKeyStatusCmd{
			Type:   pipe.Type(ctx, key),
			TTL:    pipe.TTL(ctx, key),
			Memory: pipe.MemoryUsage(ctx, key),
		}
		tmps[i] = tmp
	}
	if _, err = pipe.Exec(ctx); err != nil {
		return nil, err
	}

	for i, tmp := range tmps {
		val := response.RedisKeyMeta{
			Key:    keys[i],
			Type:   tmp.Type.Val(),
			TTL:    tmp.TTL.Val(),
			Memory: tmp.Memory.Val(),
		}

		ret.Records = append(ret.Records, val)
	}

	return ret, nil
}

func (rc *RedisConfig) Value(ctx context.Context, key string, cursor uint64, match string, count int64) (*response.RedisValue, error) {
	cli, err := rc.mem.Load(ctx)
	if err != nil {
		return nil, err
	}

	typ, err := cli.Type(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	pipe := cli.Pipeline()
	if _, err = pipe.Exec(ctx); err != nil {
		return nil, err
	}
	ret := &response.RedisValue{Key: key, Type: typ}
	switch typ {
	case "string":
		ret.Data, err = cli.Get(ctx, key).Result()

	case "hash":
		data, next, e := cli.HScan(ctx, key, cursor, match, count).Result()
		if e != nil {
			return nil, e
		}

		// HSCAN 返回 field/value 交替排列的字符串数组
		hash := make(map[string]string, len(data)/2)
		for i := 0; i+1 < len(data); i += 2 {
			hash[data[i]] = data[i+1]
		}

		ret.Data = hash
		ret.Cursor = next
	case "list":
		start := int64(cursor)
		stop := start + count - 1

		ret.Data, err = cli.LRange(ctx, key, start, stop).Result()
		ret.Cursor = uint64(start + count)
	case "set":
		data, next, e := cli.SScan(ctx, key, cursor, match, count).Result()
		if e != nil {
			return nil, e
		}

		ret.Data = data
		ret.Cursor = next
	case "zset":
		data, next, e := cli.ZScan(ctx, key, cursor, match, count).Result()
		if e != nil {
			return nil, e
		}

		ret.Data = data
		ret.Cursor = next
	case "stream":
		data, e := cli.XRangeN(ctx, key, "-", "+", count).Result()
		if e != nil {
			return nil, e
		}
		ret.Data = data
	default:
	}
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (rc *RedisConfig) Forget() {
	if cli, _ := rc.mem.Forget(); cli != nil {
		_ = cli.Close()
	}
}

func (rc *RedisConfig) SafeCommands() map[string]struct{} {
	return maps.Clone(rc.safes)
}

func (rc *RedisConfig) slowLoad(ctx context.Context) (redis.UniversalClient, error) {
	cfg, err := rc.Get(ctx)
	if err != nil {
		return nil, errcode.ErrRedisNotConfigured
	}
	if !cfg.Enabled {
		return nil, errcode.ErrRedisDisabled
	}

	cli := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:         cfg.Addresses,
		DB:            cfg.DB,
		Username:      cfg.Username,
		Password:      cfg.Password,
		MasterName:    cfg.MasterName,
		IsClusterMode: cfg.ClusterMode,
	})

	return cli, nil
}

func (rc *RedisConfig) safeCommands() map[string]struct{} {
	return map[string]struct{}{
		// Key
		"KEYS":   {},
		"EXISTS": {},
		"TYPE":   {},
		"TTL":    {},
		"PTTL":   {},
		"SCAN":   {},
		"DBSIZE": {},
		"OBJECT": {},
		"TOUCH":  {},

		// String
		"GET":      {},
		"MGET":     {},
		"GETRANGE": {},
		"STRLEN":   {},
		"GETBIT":   {},
		"BITCOUNT": {},
		"BITPOS":   {},

		// Hash
		"HGET":       {},
		"HMGET":      {},
		"HGETALL":    {},
		"HKEYS":      {},
		"HVALS":      {},
		"HLEN":       {},
		"HEXISTS":    {},
		"HSCAN":      {},
		"HRANDFIELD": {},

		// List
		"LLEN":   {},
		"LINDEX": {},
		"LRANGE": {},

		// Set
		"SCARD":       {},
		"SISMEMBER":   {},
		"SMISMEMBER":  {},
		"SMEMBERS":    {},
		"SRANDMEMBER": {},
		"SSCAN":       {},

		// ZSet
		"ZCARD":     {},
		"ZCOUNT":    {},
		"ZLEXCOUNT": {},
		"ZRANGE":    {},
		"ZRANK":     {},
		"ZREVRANK":  {},
		"ZSCORE":    {},
		"ZMSCORE":   {},
		"ZSCAN":     {},

		// Geo
		"GEOHASH":   {},
		"GEOPOS":    {},
		"GEODIST":   {},
		"GEOSEARCH": {},

		// HyperLogLog
		"PFCOUNT": {},

		// Stream
		"XLEN":      {},
		"XRANGE":    {},
		"XREVRANGE": {},
		"XINFO":     {},

		// Server
		"PING":     {},
		"ECHO":     {},
		"TIME":     {},
		"ROLE":     {},
		"LASTSAVE": {},
		"INFO":     {},

		// Other
		"MEMORY":  {},
		"SLOWLOG": {},
	}
}
