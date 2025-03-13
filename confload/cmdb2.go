package confload

import (
	"context"
	"net/url"

	"github.com/vela-ssoc/ssoc-manager/integration/cmdb2"
)

func NewCmdb2(rawURL, accessKey, secretKey string) cmdb2.Configurer {
	reqURL, err := url.Parse(rawURL)

	return &cmdb2Config{
		reqURL:    reqURL,
		accessKey: accessKey,
		secretKey: secretKey,
		err:       err,
	}
}

type cmdb2Config struct {
	reqURL    *url.URL
	accessKey string
	secretKey string
	err       error
}

func (cc *cmdb2Config) Configure(_ context.Context) (*cmdb2.Config, error) {
	if cc.err != nil {
		return nil, cc.err
	}

	reqURL := *cc.reqURL

	return &cmdb2.Config{
		URL:       &reqURL,
		AccessKey: cc.accessKey,
		SecretKey: cc.secretKey,
	}, nil
}
