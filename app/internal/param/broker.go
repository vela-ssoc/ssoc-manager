package param

type BrokerCreate struct {
	Name       string   `json:"name"       validate:"lte=20"`                                     // 名字只是为了有辨识度
	LAN        []string `json:"lan"        validate:"required_without=VIP,unique,lte=10,dive,ws"` // 内部连接地址
	VIP        []string `json:"vip"        validate:"required_without=LAN,unique,lte=10,dive,ws"` // 外部连接地址
	Bind       string   `json:"bind"       validate:"required,lte=22"`                            // 监听地址
	Servername string   `json:"servername" validate:"lte=255"`                                    // TLS 证书校验用
}

type BrokerUpdate struct {
	IntID
	BrokerCreate
}
