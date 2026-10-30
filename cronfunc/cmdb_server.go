package cronfunc

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-manager/integration/ywcmdb"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (cj *CMDB) serverAll(ctx context.Context, now time.Time) error {
	seq := cj.cli.ServerAll(ctx, nil, 500)
	errcnt := new(errorContinuous)
	for resp, err := range seq {
		if err != nil {
			cj.log.ErrorContext(ctx, "拉取 server 出错", "err", err)
			return err
		}

		err = cj.serverUpsert(ctx, resp.Result, now)
		if err != nil {
			cj.log.ErrorContext(ctx, "保存 server 出错", "err", err)
		}

		if errcnt.threshold10(err) {
			return err
		}
	}

	return nil
}

func (cj *CMDB) serverUpsert(ctx context.Context, results ywcmdb.Servers, now time.Time) error {
	size := len(results)
	cmods := make([]mongo.WriteModel, 0, size)
	for _, srv := range results {
		mod := cj.serverConvert(srv)
		cmod := cj.newCMDBBulkWriteModel(mod, now)
		cmods = append(cmods, cmod)
	}

	return cj.bulkWriteCMDB(ctx, cmods)
}

func (cj *CMDB) serverConvert(srv ywcmdb.Server) *model.CMDB {
	privateIDs := cj.splitPrivateIPs(srv.PrivateIP)
	mod := &model.CMDB{
		CIID:                    srv.ID,
		AccountID:               srv.AccountID,
		AppDuty:                 cj.dutiesConvert(srv.AppDuty),
		AppName:                 srv.CIAppName(),
		Appid:                   srv.Appid,
		AssetID:                 srv.AssetID,
		AssetStatus:             srv.AssetStatus,
		AutoRenew:               srv.AutoRenew,
		BaoleijiIdentity:        srv.BaoleijiIdentity,
		BeetleService:           srv.BeetleService,
		BillingCycle:            srv.BillingCycle,
		BillingType:             srv.BillingType,
		Business:                srv.Business,
		BusinessEnv:             srv.BusinessEnv,
		BusinessStatus:          srv.BusinessStatus,
		CPU:                     srv.CPU,
		CPUCount:                srv.CPUCount,
		CIType:                  srv.CIType,
		CmcIP:                   srv.CmcIP,
		CncIP:                   srv.CncIP,
		Comment:                 srv.Comment,
		CostDepCasID:            srv.CostDepCasID,
		CreatedTime:             srv.CreatedTime,
		CtcIP:                   srv.CtcIP,
		Deleted:                 srv.Deleted,
		Department:              srv.Department,
		Description:             srv.Description,
		DockerCPUCount:          srv.DockerCPUCount,
		Env:                     srv.Env,
		ExternalID:              srv.ExternalID,
		FloatIP:                 srv.FloatIP,
		Harddisk:                srv.Harddisk,
		HostBuyDate:             srv.HostBuyDate,
		HostIP:                  srv.HostIP,
		HostType:                srv.HostType,
		Hostname:                srv.Hostname,
		HyperThreading:          srv.HyperThreading,
		IDC:                     srv.IDC,
		IPv6:                    cj.splitPrivateIPs(srv.IPv6),
		Image:                   srv.Image,
		ImageName:               srv.ImageName,
		ImageVersion:            srv.ImageVersion,
		ImportedAt:              srv.ImportedAt,
		InScalingGroup:          srv.InScalingGroup,
		InstanceID:              srv.InstanceID,
		InstanceType:            srv.InstanceType,
		InternetMaxBandwidthOut: srv.InternetMaxBandwidthOut,
		K8SCluster:              srv.K8SCluster,
		KernelVersion:           srv.KernelVersion,
		LogicCPUCount:           srv.LogicCPUCount,
		MinionNotCheck:          srv.MinionNotCheck,
		Name:                    srv.Name,
		Namespace:               srv.Namespace,
		NetOpen:                 srv.NetOpen,
		OSArch:                  srv.OSArch,
		OSName:                  srv.OSName,
		OSType:                  srv.OSType,
		OSVersion:               srv.OSVersion,
		OpDuty:                  cj.dutiesConvert(srv.OpDutyMain),
		Payer:                   srv.Payer,
		PowerStates:             srv.PowerStates,
		PrivateCloudIP:          srv.PrivateCloudIP,
		PrivateCloudType:        srv.PrivateCloudType,
		PrivateIP:               privateIDs,
		PublicCloudID:           srv.PublicCloudID,
		PublicCloudIDC:          srv.PublicCloudIDC,
		PublicCloudCreatedAt:    srv.CreatedAt,
		PublicCloudExpiredAt:    srv.ExpiredAt,
		PublicCloudUpdatedAt:    srv.UpdatedAt,
		RAID:                    srv.RAID,
		RAM:                     srv.RAM,
		RAMSize:                 srv.RAMSize,
		Rack:                    srv.Rack,
		RdDuty:                  cj.dutiesConvert(srv.RdDutyMain),
		Region:                  srv.Region,
		RegionExtID:             srv.RegionExtID,
		ResourceRequests:        srv.ResourceRequests,
		SN:                      srv.SN,
		SecurityInfo:            srv.SecurityInfo,
		Server:                  srv.Server,
		ServerRoom:              srv.ServerRoom,
		ShutdownBehavior:        srv.ShutdownBehavior,
		ShutdownMode:            srv.ShutdownMode,
		Status:                  srv.Status,
		SysDuty:                 cj.dutiesConvert(srv.SysDuty),
		Tags:                    srv.Tags,
		Throughput:              srv.Throughput,
		TradeType:               srv.TradeType,
		UUID:                    srv.UUID,
		UpdateTime:              srv.UpdateTime,
		Use:                     srv.Use,
		VCPUCount:               srv.VCPUCount,
		VMemSize:                srv.VMemSize,
		VServerType:             srv.VServerType,
		ZabbixNotCheck:          srv.ZabbixNotCheck,
		Zone:                    srv.Zone,
	}

	return mod
}
