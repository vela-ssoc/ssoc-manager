package ywcmdb

type PublicCloudEIP struct {
	ID               int64  `json:"_id,string"`
	AddressType      string `json:"address_type,omitempty"`
	BindInstanceID   string `json:"bind_instance_id,omitempty"`
	CIType           string `json:"ci_type"`
	CloudCreatedTime string `json:"cloud_created_time"`
	CloudProduct     string `json:"cloud_product"`
	CreatedTime      string `json:"created_time"`
	Department       string `json:"department"`
	ImportedAt       string `json:"imported_at"`
	InstanceID       string `json:"instance_id"`
	IP               string `json:"ip,omitempty"`
	OpDuty           Duties `json:"op_duty,omitempty"`
	Payer            string `json:"payer"`
	PublicCloudIDC   string `json:"public_cloud_idc"`
	Region           string `json:"region"`
	UpdateTime       string `json:"update_time"`
	Zone             string `json:"zone,omitempty"`
	IPv6             string `json:"ipv6,omitempty"`
	BandwidthID      string `json:"bandwidth_id,omitempty"`
	IPVersion        int    `json:"ip_version,omitempty"`
	ProjectID        string `json:"project_id,omitempty"`
	ProjectName      string `json:"project_name,omitempty"`
}

type PublicCloudEIPs []PublicCloudEIP

// PublicCloudLB 公有云负载均衡
type PublicCloudLB struct {
	Id              string `json:"_id"`
	AddressType     string `json:"address_type"`
	BackendInfoList []struct {
		BackendType   string  `json:"backend_type"`
		InstanceId    *string `json:"instance_id"`
		Ip            *string `json:"ip"`
		ServerGroupId string  `json:"server_group_id,omitempty"`
		Weight        int     `json:"weight,omitempty"`
		Ipv6          *string `json:"ipv6,omitempty"`
	} `json:"backend_info_list,omitempty"`
	CiType           string   `json:"ci_type"`
	CloudCreatedTime string   `json:"cloud_created_time"`
	CloudProduct     string   `json:"cloud_product"`
	CreatedTime      string   `json:"created_time"`
	Department       string   `json:"department"`
	ImportedAt       string   `json:"imported_at"`
	InstanceId       string   `json:"instance_id"`
	Ip               []string `json:"ip,omitempty"`
	ListenInfoList   []struct {
		Enable        bool     `json:"enable"`
		ForwardPort   *int     `json:"forward_port,omitempty"`
		ListenPort    int      `json:"listen_port"`
		Protocol      string   `json:"protocol"`
		ServerGroupId []string `json:"server_group_id,omitempty"`
	} `json:"listen_info_list"`
	Payer          string   `json:"payer"`
	PublicCloudIdc string   `json:"public_cloud_idc"`
	Region         string   `json:"region"`
	UpdateTime     string   `json:"update_time"`
	Zones          []string `json:"zones,omitempty"`
	OpDuty         []struct {
		Block      bool   `json:"block"`
		EmployeeId string `json:"employee_id"`
		Nickname   string `json:"nickname"`
		Uid        int    `json:"uid"`
		Username   string `json:"username"`
		Utype      string `json:"utype"`
	} `json:"op_duty,omitempty"`
	ProjectId       string   `json:"project_id,omitempty"`
	Ipv6            []string `json:"ipv6,omitempty"`
	ServiceInfoList []struct {
		ClusterName string  `json:"cluster_name"`
		Namespace   string  `json:"namespace"`
		NodePort    *string `json:"node_port"`
		Port        string  `json:"port"`
		Protocol    string  `json:"protocol"`
		ServiceName string  `json:"service_name"`
		TargetPort  string  `json:"target_port"`
	} `json:"service_info_list,omitempty"`
	IpIntra        []string `json:"ip_intra,omitempty"`
	Ipv6Intra      []string `json:"ipv6_intra,omitempty"`
	Ipv6Mode       string   `json:"ipv6_mode,omitempty"`
	Ipv6ManualMark string   `json:"ipv6_manual_mark,omitempty"`
}
