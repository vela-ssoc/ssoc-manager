package request

type SystemPID struct {
	PID int32 `json:"pid" query:"pid" form:"pid" param:"pid" jsonschema:"required" jsonschema_description:"进程 PID"`
}

type SystemNetworkConnection struct {
	Kind string `json:"kind" jsonschema:"enum=all,enum=tcp,enum=udp" jsonschema_description:"网络连接类型，all 表示所有类型"`
}
