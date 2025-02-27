package cmdb2

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/vela-ssoc/vela-common-mb/httpx"
)

var (
	ciTypeServers = []string{"server", "vserver", "vmware_vserver", "public_cloud", "docker"}
	ciTypeVIPs    = []string{"vip"}
)

type Config struct {
	URL       *url.URL
	AccessKey string
	SecretKey string
}

type Configurer interface {
	Configure(ctx context.Context) (*Config, error)
}

type Client interface {
	Servers(ctx context.Context, ips []string, page, size int) ([]*Server, error)
	VIPs(ctx context.Context, page, size int) ([]*VIP, error)
}

func NewClient(cfg Configurer, cli httpx.Client) Client {
	return cmdb2Client{
		cfg: cfg,
		cli: cli,
	}
}

type cmdb2Client struct {
	cfg Configurer
	cli httpx.Client
}

func (c cmdb2Client) Servers(ctx context.Context, ips []string, page, size int) ([]*Server, error) {
	if size <= 0 {
		size = 30
	}

	ret := make([]*Server, 0, size)
	if err := c.getCITypes(ctx, ciTypeServers, ips, page, size, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (c cmdb2Client) VIPs(ctx context.Context, page, size int) ([]*VIP, error) {
	if size <= 0 {
		size = 30
	}

	ret := make([]*VIP, 0, size)
	if err := c.getCITypes(ctx, ciTypeVIPs, nil, page, size, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (c cmdb2Client) getCITypes(ctx context.Context, ciTypes, ips []string, page, size int, ret any) error {
	types := strings.Join(ciTypes, ";")
	args := "_type:(" + types + ")"
	if len(ips) != 0 {
		inets := strings.Join(ips, ";")
		args = args + ",private_ip:(" + inets + ")"
	}
	quires := url.Values{"q": []string{args}}

	if page > 0 {
		quires.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		quires.Set("count", strconv.Itoa(size))
	}

	return c.getCI(ctx, quires, ret)
}

func (c cmdb2Client) getCI(ctx context.Context, quires url.Values, ret any) error {
	cfg, err := c.cfg.Configure(ctx)
	if err != nil {
		return err
	}
	reqURL := cfg.URL.JoinPath("/cmdb2/api/v0.1/ci/s")
	query := reqURL.Query()
	query.Set("_key", cfg.AccessKey)
	query.Set("_secret", cfg.SecretKey)
	for k, vs := range quires {
		query[k] = vs
	}
	reqURL.RawQuery = query.Encode()
	strURL := reqURL.String()

	return c.cli.JSON(ctx, strURL, nil, &result{Records: ret})
}
