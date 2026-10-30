package ywcmdb

import (
	"context"
	"iter"
	"log/slog"
	"net/http"
)

type Client interface {
	ServerAll(ctx context.Context, ciTypes []string, size int) iter.Seq2[*ResponseBody[Servers, Server], error]

	DepartmentAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[Departments, Department], error]

	PrivateDNSRecordAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[PrivateDNSRecords, PrivateDNSRecord], error]

	PublicDNSRecordAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[PublicDNSRecords, PublicDNSRecord], error]

	PublicCloudEIPAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[PublicCloudEIPs, PublicCloudEIP], error]

	VIPAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[VIPs, VIP], error]
}

type Config struct {
	// Endpoint 请求地址。
	// https://yunwei.example.com/cmdb2/api/v0.1/ci/s?foo=bar
	Endpoint string
	Key      string `json:"key"`
	Secret   string `json:"secret"`
}

type ConfigLoader interface {
	LoadCMDBConfig(ctx context.Context) (*Config, error)
}

type cmdbClient struct {
	cfg ConfigLoader
	cli *http.Client
	log *slog.Logger
}

func NewClient(cfg ConfigLoader, cli *http.Client, log *slog.Logger) Client {
	return &cmdbClient{
		cfg: cfg,
		cli: cli,
		log: log,
	}
}

func (cc *cmdbClient) ServerAll(ctx context.Context, ciTypes []string, size int) iter.Seq2[*ResponseBody[Servers, Server], error] {
	if len(ciTypes) == 0 {
		ciTypes = []string{"server", "vserver", "fund_vserver", "public_cloud", "docker"}
	}
	sub := newSubClient[Servers, Server](ctx, ciTypes, size, cc)

	return sub.All()
}

func (cc *cmdbClient) DepartmentAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[Departments, Department], error] {
	ciTypes := []string{"department"}
	sub := newSubClient[Departments, Department](ctx, ciTypes, size, cc)

	return sub.All()
}

func (cc *cmdbClient) PrivateDNSRecordAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[PrivateDNSRecords, PrivateDNSRecord], error] {
	ciTypes := []string{"private_dns_record"}
	sub := newSubClient[PrivateDNSRecords, PrivateDNSRecord](ctx, ciTypes, size, cc)

	return sub.All()
}

func (cc *cmdbClient) PublicDNSRecordAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[PublicDNSRecords, PublicDNSRecord], error] {
	ciTypes := []string{"public_dns_record"}
	sub := newSubClient[PublicDNSRecords, PublicDNSRecord](ctx, ciTypes, size, cc)

	return sub.All()
}

func (cc *cmdbClient) PublicCloudEIPAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[PublicCloudEIPs, PublicCloudEIP], error] {
	ciTypes := []string{"public_cloud_eip"}
	sub := newSubClient[PublicCloudEIPs, PublicCloudEIP](ctx, ciTypes, size, cc)

	return sub.All()
}

func (cc *cmdbClient) VIPAll(ctx context.Context, size int) iter.Seq2[*ResponseBody[VIPs, VIP], error] {
	ciTypes := []string{"vip"}
	sub := newSubClient[VIPs, VIP](ctx, ciTypes, size, cc)

	return sub.All()
}
