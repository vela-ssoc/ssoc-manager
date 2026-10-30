package ywcmdb

// Server CMDB 服务器资产。
//
// 物理机：https://oa-pan.eastmoney.com/ddwiki/space/doc?spaceId=15&fileUuid=9e7fb369-3695-406d-85ae-aa51579f94aa
// 虚拟机：https://oa-pan.eastmoney.com/ddwiki/space/doc?spaceId=15&fileUuid=20ab7bdd-beeb-4031-bb46-ec40986592e3
// 容器：https://oa-pan.eastmoney.com/ddwiki/space/doc?spaceId=15&fileUuid=3ce158ee-31f8-4755-b7ad-0fb5c1a24ffb
// 公有云：https://oa-pan.eastmoney.com/ddwiki/space/doc?spaceId=15&fileUuid=44bd2034-a9b1-4c74-9b25-0c147783f4e4
//
// 不同类型（ci_type）的唯一标识：
//   - server:       sn (SN)
//   - vserver:      uuid (UUID)
//   - docker:       instance_id (InstanceID)
//   - public_cloud: id (PublicCloudID)
type Server struct {
	ID                      int64    `json:"_id,string"`
	AccountID               string   `json:"account_id"`
	AppDuty                 Duties   `json:"app_duty,omitzero"`
	AppName                 string   `json:"app_name,omitzero"` // 机器所属的应用名称，注意和 appname 字段的区别
	Appname                 string   `json:"appname,omitzero"`  // 对应老版 CMDB 的应用名字段
	Appid                   string   `json:"appid"`
	AssetID                 string   `json:"asset_id,omitzero"`
	AssetStatus             string   `json:"asset_status,omitzero"`
	AutoRenew               string   `json:"auto_renew"`
	BaoleijiIdentity        string   `json:"baoleiji_identity,omitzero"`
	BeetleService           string   `json:"beetle_service,omitzero"`
	BillingCycle            string   `json:"billing_cycle,omitzero"`
	BillingType             string   `json:"billing_type"`
	Business                string   `json:"business,omitzero"`
	BusinessEnv             string   `json:"business_env"`
	BusinessStatus          string   `json:"business_status,omitzero"`
	CPU                     string   `json:"cpu,omitzero"`
	CPUCount                int      `json:"cpu_count,omitzero"`
	CIType                  string   `json:"ci_type"`
	CmcIP                   string   `json:"cmc_ip,omitzero"`
	CncIP                   string   `json:"cnc_ip,omitzero"`
	Comment                 string   `json:"comment"`
	CostDepCasID            string   `json:"cost_dep_cas_id,omitzero"`
	CreatedAt               string   `json:"created_at"`
	CreatedTime             string   `json:"created_time"`
	CtcIP                   string   `json:"ctc_ip,omitzero"`
	Deleted                 string   `json:"deleted"`
	Department              string   `json:"department"`
	Description             string   `json:"description,omitzero"`
	DockerCPUCount          string   `json:"docker_cpu_count"`
	Env                     string   `json:"env"`
	ExpiredAt               string   `json:"expired_at,omitzero"`
	ExternalID              string   `json:"external_id"`
	FloatIP                 []string `json:"float_ip,omitzero"`
	Harddisk                string   `json:"harddisk,omitzero"`
	HostBuyDate             string   `json:"host_buy_date,omitzero"`
	HostIP                  []string `json:"host_ip"`
	HostType                string   `json:"host_type,omitzero"`
	Hostname                string   `json:"hostname,omitzero"`
	HyperThreading          string   `json:"hyper_threading,omitzero"`
	IDC                     string   `json:"idc"`
	IPv6                    []string `json:"ipv6,omitzero"`
	Image                   string   `json:"image"`
	ImageName               string   `json:"image_name,omitzero"`
	ImageVersion            string   `json:"image_version"`
	ImportedAt              string   `json:"imported_at"`
	InScalingGroup          string   `json:"in_scaling_group"`
	InstanceID              string   `json:"instance_id"`
	InstanceType            string   `json:"instance_type"`
	InternetMaxBandwidthOut int      `json:"internet_max_bandwidth_out,omitzero"`
	K8SCluster              string   `json:"k8s_cluster"`
	KernelVersion           string   `json:"kernel_version,omitzero"`
	LogicCPUCount           int      `json:"logic_cpu_count,omitzero"`
	MinionNotCheck          string   `json:"minion_not_check,omitzero"`
	Name                    string   `json:"name"`
	Namespace               string   `json:"namespace"`
	NetOpen                 string   `json:"net_open,omitzero"`
	OSArch                  string   `json:"os_arch"`
	OSName                  string   `json:"os_name"`
	OSType                  string   `json:"os_type"`
	OSVersion               string   `json:"os_version,omitzero"`
	OpDuty                  DutyMain `json:"op_duty"`
	OpDutyMain              Duties   `json:"op_duty.main"`
	Payer                   string   `json:"payer,omitzero"`
	PowerStates             string   `json:"power_states"`
	PrivateCloudIP          string   `json:"private_cloud_ip,omitzero"`
	PrivateCloudType        string   `json:"private_cloud_type,omitzero"`
	PrivateIP               []string `json:"private_ip"`
	PublicCloudID           string   `json:"id"` // public_cloud 云管 ID
	PublicCloudIDC          string   `json:"public_cloud_idc"`
	RAID                    string   `json:"raid,omitzero"`
	RAM                     string   `json:"ram,omitzero"`
	RAMSize                 string   `json:"ram_size,omitzero"`
	Rack                    string   `json:"rack"`
	RdDuty                  DutyMain `json:"rd_duty,omitzero"`
	RdDutyMain              Duties   `json:"rd_duty.main,omitzero"`
	Region                  string   `json:"region"`
	RegionExtID             string   `json:"region_ext_id"`
	ResourceRequests        string   `json:"resource_requests"`
	SN                      string   `json:"sn"`
	SecurityInfo            string   `json:"security_info,omitzero"`
	Server                  string   `json:"server"`
	ServerRoom              string   `json:"server_room"`
	ShutdownBehavior        string   `json:"shutdown_behavior"`
	ShutdownMode            string   `json:"shutdown_mode,omitzero"`
	Status                  string   `json:"status"`
	SysDuty                 Duties   `json:"sys_duty,omitzero"`
	Tags                    []string `json:"tags,omitzero"`
	Throughput              int      `json:"throughput,omitzero"`
	TradeType               string   `json:"trade_type,omitzero"`
	UUID                    string   `json:"uuid"`
	UpdateTime              string   `json:"update_time"`
	UpdatedAt               string   `json:"updated_at,omitzero"` // 云管更新时间 YYYY-mm-DD HH:MM:SS
	Use                     string   `json:"use,omitzero"`
	VCPUCount               int      `json:"vcpu_count"`
	VMemSize                int      `json:"vmem_size"`
	VServerType             string   `json:"vserver_type"`
	ZabbixNotCheck          string   `json:"zabbix_not_check,omitzero"`
	Zone                    string   `json:"zone"`

	// AppLoadType             []string `json:"app_load_type,omitzero"`
	// SecurityRisk            int      `json:"security_risk,omitzero"`
}

type Servers []Server

// CIAppName 不同类型（ci_type）的唯一标识：
//
//   - server:       sn (SN)
//   - vserver:      uuid (UUID)
//   - docker:       instance_id (InstanceID)
//   - public_cloud: id (PublicCloudID)
func (s Server) CIAppName() string {
	if s.AppName != "" {
		return s.AppName
	}

	return s.Appname
}

func (s Server) CILocation() string {
	switch s.CIType {
	case "public_cloud":
		return s.PublicCloudIDC
	default:
		return s.IDC
	}
}
