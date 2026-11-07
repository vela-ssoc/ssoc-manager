package request

type RedisConfigUpsert struct {
	Enabled     bool     `json:"enabled"`
	Addresses   []string `json:"addresses"     validate:"required,gte=1,lte=100,dive,required"` // 集群模式填写多个，哨兵模式填写哨兵地址
	Username    string   `json:"username"`                                                      // Redis 6+
	Password    string   `json:"password"      validate:"lte=100"`                              // 密码
	ClusterMode bool     `json:"cluster_mode"`                                                  // 是否集群模式
	MasterName  string   `json:"master_name"   validate:"lte=100"`                              // 哨兵模式必填
	DB          int      `json:"db"            validate:"gte=0,lte=15"`
}

type RedisInfo struct {
	Section []string `json:"section" query:"section" validate:"lte=100,unique,dive,required"`
}

type RedisDo struct {
	Args      []string `json:"args"      validate:"gte=1,lte=20,dive,required"`
	Privilege bool     `json:"privilege"`
}

func (rd RedisDo) AnyArgs() []any {
	ret := make([]any, 0, len(rd.Args))
	for _, arg := range rd.Args {
		ret = append(ret, arg)
	}

	return ret
}
