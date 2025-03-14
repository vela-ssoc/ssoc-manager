package mapstruct

import (
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-manager/integration/cmdb2"
)

func Cmdb2Server(d *cmdb2.Server) *model.Cmdb2 {
	var opDuty model.Cmdb2OpDuty
	opDuty.Main = Cmdb2Duties(d.OpDuty.Main)
	opDuty.Backup = Cmdb2Duties(d.OpDuty.Backup)
	opDuty.Standby = Cmdb2Duties(d.OpDuty.Standby)

	return &model.Cmdb2{
		AppCluster:              d.AppCluster,
		AppDuty:                 Cmdb2Duties(d.AppDuty),
		AppID:                   d.AppID,
		AppName:                 d.AppName,
		AutoRenew:               d.AutoRenew,
		AssetID:                 d.AssetID,
		AssetStatus:             d.AssetStatus,
		BaoleijiIdentity:        d.BaoleijiIdentity,
		BeetleService:           d.BeetleService,
		BillingType:             d.BillingType,
		Brand:                   d.Brand,
		Business:                d.Business,
		BusinessEnv:             d.BusinessEnv,
		ChargeMode:              d.ChargeMode,
		CIType:                  d.CIType,
		CmcIP:                   d.CmcIP,
		CncIP:                   d.CncIP,
		Comment:                 d.Comment,
		CostDepCasID:            d.CostDepCasID,
		CPU:                     d.CPU,
		CPUCount:                d.CPUCount,
		CtcIP:                   d.CtcIP,
		CreateDate:              d.CreateDate,
		CreatedAt:               d.CreatedAt,
		CreatedTime:             d.CreatedTime,
		Deleted:                 d.Deleted,
		Department:              d.Department,
		Description:             d.Description,
		DeviceSpec:              d.DeviceSpec,
		DockerCPUCount:          d.DockerCPUCount,
		Env:                     d.Env,
		ExpDate:                 d.ExpDate,
		ExpiredAt:               d.ExpiredAt,
		ExternalID:              d.ExternalID,
		FloatIP:                 d.FloatIP,
		HardDisk:                d.HardDisk,
		HostIP:                  d.HostIP,
		HostType:                d.HostType,
		Hostname:                d.Hostname,
		HostSN:                  d.HostSN,
		HyperThreading:          d.HyperThreading,
		IDC:                     d.IDC,
		IloIP:                   d.IloIP,
		Image:                   d.Image,
		ImageVersion:            d.ImageVersion,
		InstanceID:              d.InstanceID,
		IPv6:                    d.IPv6,
		ImportedAt:              d.ImportedAt,
		InScalingGroup:          d.InScalingGroup,
		InstanceType:            d.InstanceType,
		InternetMaxBandwidthOut: d.InternetMaxBandwidthOut,
		K8sCluster:              d.K8sCluster,
		KernelVersion:           d.KernelVersion,
		LogicCPUCount:           d.LogicCPUCount,
		MinionNotCheck:          d.MinionNotCheck,
		Name:                    d.Name,
		Namespace:               d.Namespace,
		NetOpen:                 d.NetOpen,
		OpDuty:                  opDuty,
		OpDutyBackup:            Cmdb2Duties(d.OpDutyBackup),
		OpDutyMain:              Cmdb2Duties(d.OpDutyMain),
		OpDutyStandby:           Cmdb2Duties(d.OpDutyStandby),
		OsArch:                  d.OsArch,
		OsType:                  d.OsType,
		OsVersion:               d.OsVersion,
		PowerStates:             d.PowerStates,
		PrivateIP:               d.PrivateIP,
		PublicCloudID:           d.PublicCloudID,
		PublicCloudIDC:          d.PublicCloudIDC,
		PrivateCloudIP:          d.PrivateCloudIP,
		PrivateCloudType:        d.PrivateCloudType,
		Rack:                    d.Rack,
		RAID:                    d.RAID,
		RAM:                     d.RAM,
		RAMSize:                 d.RAMSize,
		RdDutyMain:              Cmdb2Duties(d.RdDutyMain),
		RdDutyMember:            Cmdb2Duties(d.RdDutyMember),
		Region:                  d.Region,
		ResourceLimits:          d.ResourceLimits,
		ResourceRequests:        d.ResourceRequests,
		SecurityInfo:            d.SecurityInfo,
		Server:                  d.Server,
		ServerRoom:              d.ServerRoom,
		SN:                      d.SN,
		ShutdownBehavior:        d.ShutdownBehavior,
		ShutdownMode:            d.ShutdownMode,
		Status:                  d.Status,
		SysDuty:                 Cmdb2Duties(d.SysDuty),
		Tags:                    d.Tags,
		Throughput:              d.Throughput,
		TradeType:               d.TradeType,
		UpdateTime:              d.UpdateTime,
		UpdatedAt:               d.UpdatedAt,
		Use:                     d.Use,
		UUID:                    d.UUID,
		VCPUCount:               d.VCPUCount,
		VMemSize:                d.VMemSize,
		VserverType:             d.VserverType,
		ZabbixNotCheck:          d.ZabbixNotCheck,
		Zone:                    d.Zone,
	}
}

func Cmdb2Duties(dats []*cmdb2.DutyMain) model.Cmdb2Duties {
	duties := make(model.Cmdb2Duties, 0, len(dats))
	for _, d := range dats {
		du := &model.Cmdb2Duty{
			EmployeeID: d.EmployeeID,
			Nickname:   d.Nickname,
			UID:        d.UID,
			Username:   d.Username,
			UType:      d.UType,
		}
		duties = append(duties, du)
	}

	return duties
}
