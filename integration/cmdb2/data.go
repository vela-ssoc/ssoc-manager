package cmdb2

import (
	"net"
	"strconv"
)

type result struct {
	NumFound int `json:"numFound"`
	Numfound int `json:"numfound"`
	Page     int `json:"page"`
	Total    int `json:"total"`
	Records  any `json:"result"`
}

type OpDuty struct {
	Backup  []*DutyMain `json:"backup,omitempty"`
	Main    []*DutyMain `json:"main,omitempty"`
	Standby []*DutyMain `json:"standby,omitempty"`
}

type DutyMain struct {
	EmployeeID string `json:"employee_id,omitempty"`
	Nickname   string `json:"nickname,omitempty"`
	UID        int    `json:"uid,omitempty"`
	Username   string `json:"username,omitempty"`
	UType      string `json:"utype,omitempty"`
}

type Server struct {
	AppCluster              string      `json:"app_cluster,omitempty"`
	AppDuty                 []*DutyMain `json:"app_duty,omitempty"`
	AppID                   string      `json:"appid,omitempty"`
	AppName                 string      `json:"appname,omitempty"`
	AutoRenew               string      `json:"auto_renew,omitempty"`
	AssetID                 string      `json:"asset_id,omitempty"`
	AssetStatus             string      `json:"asset_status,omitempty"`
	BaoleijiIdentity        string      `json:"baoleiji_identity,omitempty"`
	BeetleService           string      `json:"beetle_service,omitempty"`
	BillingType             string      `json:"billing_type,omitempty"`
	Brand                   string      `json:"brand,omitempty"`
	Business                string      `json:"business,omitempty"`
	BusinessEnv             string      `json:"business_env,omitempty"`
	ChargeMode              string      `json:"charge_mode,omitempty"`
	CIType                  string      `json:"ci_type,omitempty"`
	CmcIP                   string      `json:"cmc_ip,omitempty"`
	CncIP                   string      `json:"cnc_ip,omitempty"`
	Comment                 string      `json:"comment,omitempty"`
	CostDepCasID            string      `json:"cost_dep_cas_id,omitempty"`
	CPU                     string      `json:"cpu,omitempty"`
	CPUCount                int         `json:"cpu_count,omitempty"`
	CtcIP                   string      `json:"ctc_ip,omitempty"`
	CreateDate              string      `json:"create_date,omitempty"`
	CreatedAt               string      `json:"created_at,omitempty"`
	CreatedTime             string      `json:"created_time,omitempty"`
	Deleted                 string      `json:"deleted,omitempty"`
	Department              string      `json:"department,omitempty"`
	Description             string      `json:"description,omitempty"`
	DeviceSpec              string      `json:"device_spec,omitempty"`
	DockerCPUCount          string      `json:"docker_cpu_count,omitempty"`
	Env                     string      `json:"env,omitempty"`
	ExpDate                 string      `json:"exp_date,omitempty"`
	ExpiredAt               string      `json:"expired_at,omitempty"`
	ExternalID              string      `json:"external_id,omitempty"`
	FloatIP                 []string    `json:"float_ip,omitempty"`
	HardDisk                string      `json:"harddisk,omitempty"`
	HostIP                  []string    `json:"host_ip,omitempty"`
	HostType                string      `json:"host_type,omitempty"`
	Hostname                string      `json:"hostname,omitempty"`
	HostSN                  string      `json:"host_sn,omitempty"`
	HyperThreading          string      `json:"hyper_threading,omitempty"`
	IDC                     string      `json:"idc,omitempty"`
	IloIP                   string      `json:"ilo_ip,omitempty"`
	Image                   string      `json:"image,omitempty"`
	ImageVersion            string      `json:"image_version,omitempty"`
	InstanceID              string      `json:"instance_id,omitempty"`
	IPv6                    []string    `json:"ipv6,omitempty"`
	ImportedAt              string      `json:"imported_at,omitempty"`
	InScalingGroup          string      `json:"in_scaling_group,omitempty"`
	InstanceType            string      `json:"instance_type,omitempty"`
	InternetMaxBandwidthOut int         `json:"internet_max_bandwidth_out,omitempty"`
	K8sCluster              string      `json:"k8s_cluster,omitempty"`
	KernelVersion           string      `json:"kernel_version,omitempty"`
	LogicCPUCount           int         `json:"logic_cpu_count,omitempty"`
	MinionNotCheck          string      `json:"minion_not_check,omitempty"`
	Name                    string      `json:"name,omitempty"`
	Namespace               string      `json:"namespace,omitempty"`
	NetOpen                 string      `json:"net_open,omitempty"`
	OpDuty                  OpDuty      `json:"op_duty,omitempty"`
	OpDutyBackup            []*DutyMain `json:"op_duty.backup,omitempty"`
	OpDutyMain              []*DutyMain `json:"op_duty.main,omitempty"`
	OpDutyStandby           []*DutyMain `json:"op_duty.standby,omitempty"`
	OsArch                  string      `json:"os_arch,omitempty"`
	OsType                  string      `json:"os_type,omitempty"`
	OsVersion               string      `json:"os_version,omitempty"`
	PowerStates             string      `json:"power_states,omitempty"`
	PrivateIP               []string    `json:"private_ip,omitempty"`
	PublicCloudID           string      `json:"id,omitempty"` // 云管 id
	PublicCloudIDC          string      `json:"public_cloud_idc,omitempty"`
	PrivateCloudIP          string      `json:"private_cloud_ip,omitempty"`
	PrivateCloudType        string      `json:"private_cloud_type,omitempty"`
	Rack                    string      `json:"rack,omitempty"`
	RAID                    string      `json:"raid,omitempty"`
	RAM                     string      `json:"ram,omitempty"`
	RAMSize                 string      `json:"ram_size,omitempty"`
	RdDutyMain              []*DutyMain `json:"rd_duty.main,omitempty"`
	RdDutyMember            []*DutyMain `json:"rd_duty.member,omitempty"`
	Region                  string      `json:"region,omitempty"`
	ResourceLimits          string      `json:"resource_limits,omitempty"`
	ResourceRequests        string      `json:"resource_requests,omitempty"`
	SecurityInfo            string      `json:"security_info,omitempty"`
	Server                  string      `json:"server,omitempty"`
	ServerRoom              string      `json:"server_room,omitempty"`
	SN                      string      `json:"sn,omitempty"`
	ShutdownBehavior        string      `json:"shutdown_behavior,omitempty"`
	ShutdownMode            string      `json:"shutdown_mode,omitempty"`
	Status                  string      `json:"status,omitempty"`
	SysDuty                 []*DutyMain `json:"sys_duty,omitempty"`
	Tags                    []string    `json:"tags,omitempty"`
	Throughput              int         `json:"throughput,omitempty"`
	TradeType               string      `json:"trade_type,omitempty"`
	UpdateTime              string      `json:"update_time,omitempty"`
	UpdatedAt               string      `json:"updated_at,omitempty"`
	Use                     string      `json:"use,omitempty"`
	UUID                    string      `json:"uuid,omitempty"`
	VCPUCount               int         `json:"vcpu_count,omitempty"`
	VMemSize                int         `json:"vmem_size,omitempty"`
	VserverType             string      `json:"vserver_type,omitempty"`
	ZabbixNotCheck          string      `json:"zabbix_not_check,omitempty"`
	Zone                    string      `json:"zone,omitempty"`

	// 下面的字段定义和 cmdb2 接口文档描述不一致，或者是会返回多种类型。
	// 这些字段无关紧要，目前为了不影响系统间交互，就不解析存储这些字段。
	// cmdb 系统部分字段定义比较混乱，同一个字段，返回的类型都可能不一样。
	// AppLoadType      []string           `bson:"app_load_type,omitempty"      json:"app_load_type,omitempty"`
	// SecurityRisk     int                `bson:"security_risk,omitempty"      json:"security_risk,omitempty"`
}

func (s Server) Duties() []*DutyMain {
	diff := make(map[string]struct{}, 8)
	duties := make([]*DutyMain, 0, 8)

	putFunc := func(source []*DutyMain) {
		for _, d := range source {
			uname := d.Username
			if _, exist := diff[uname]; !exist {
				diff[uname] = struct{}{}
				duties = append(duties, d)
			}
		}
	}

	putFunc(s.AppDuty)
	putFunc(s.OpDuty.Main)
	putFunc(s.OpDuty.Backup)
	putFunc(s.OpDuty.Standby)
	putFunc(s.OpDutyMain)
	putFunc(s.OpDutyBackup)
	putFunc(s.OpDutyStandby)
	putFunc(s.RdDutyMain)
	putFunc(s.RdDutyMember)
	putFunc(s.SysDuty)

	return duties
}

// VIP cmdb2 虚拟 IP 资产。
//
// https://oa-pan.eastmoney.com/ddwiki/space/doc?spaceId=15&fileUuid=c292a18c-09b9-47b8-b364-7f4b31dc6f24
type VIP struct {
	CID           string        `json:"_id,omitempty"`
	AccessControl string        `json:"access_control,omitempty"`
	Certificate   string        `json:"certificate,omitempty"`
	Cost          string        `json:"cost,omitempty"`
	Department    string        `json:"department,omitempty"`
	IDC           string        `json:"idc,omitempty"`
	IdleTimeout   string        `json:"idle_timeout,omitempty"`
	IPType        string        `json:"ip_type,omitempty"`
	LoadBalance   string        `json:"load_balance,omitempty"`
	Protocol      string        `json:"protocol,omitempty"`
	RServer       []*VIPRServer `json:"rserver,omitempty"`
	Source        string        `json:"source,omitempty"`
	Status        string        `json:"status,omitempty"`
	VIP           string        `json:"vip,omitempty"`
	VIPPort       string        `json:"vip_port,omitempty"`
	VPort         int           `json:"vport,omitempty"`
	GetClientIP   string        `json:"get_client_ip,omitempty"`
}

func (v VIP) PrimaryKey() string {
	port := strconv.Itoa(v.VPort)
	return net.JoinHostPort(v.VIP, port)
}

type VIPRServer struct {
	Port   int    `json:"port,omitempty"`
	Server string `json:"server,omitempty"`
	Status string `json:"status,omitempty"`
}
