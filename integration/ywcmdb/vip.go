package ywcmdb

type VIP struct {
	ID            int64        `json:"_id,string"`
	AccessControl string       `json:"access_control,omitempty"`
	CIType        string       `json:"ci_type"`
	CostDepCasID  string       `json:"cost_dep_cas_id"`
	Department    string       `json:"department"`
	DeviceName    string       `json:"device_name"`
	IDC           string       `json:"idc"`
	IdleTimeout   string       `json:"idle_timeout"`
	IPType        string       `json:"ip_type"`
	ISP           string       `json:"isp,omitempty"`
	LoadBalance   string       `json:"load_balance"`
	NAT           int          `json:"nat"`
	Protocol      string       `json:"protocol"`
	RServer       []VIPRServer `json:"rserver"`
	ServerRoom    string       `json:"server_room"`
	Source        string       `json:"source"`
	Status        string       `json:"status"`
	VIP           string       `json:"vip"`
	VIPUniqueID   string       `json:"vip_unique_id"`
	VPort         int          `json:"vport"`
	VSName        string       `json:"vs_name"`
	Duty          Duties       `json:"duty,omitempty"`
	GetClientIP   string       `json:"get_client_ip,omitempty"`
	Certificate   string       `json:"certificate,omitempty"`
}

type VIPs []VIP

type VIPRServer struct {
	Port       int    `json:"port"`
	Server     string `json:"server"`
	Status     string `json:"status"`
	UpdateTime string `json:"update_time"`
}
