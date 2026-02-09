package brkclient

import (
	"context"
	"net/http"

	"github.com/vela-ssoc/ssoc-common/tundata/mbreq"
	"github.com/vela-ssoc/ssoc-common/tundata/mbresp"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
	"github.com/vela-ssoc/ssoc-proto/muxtool"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Client struct {
	base muxtool.Client
}

func NewClient(base muxtool.Client) Client {
	return Client{base: base}
}

func (c Client) Base() muxtool.Client {
	return c.base
}

func (c Client) TunnelStat(ctx context.Context, brokerID bson.ObjectID) (*mbresp.TunnelStat, error) {
	reqURL := muxproto.ManagerToBrokerURL(brokerID.Hex(), "/api/v1/tunnel/stat")
	strURL := reqURL.String()

	ret := new(mbresp.TunnelStat)
	if err := c.base.JSON(ctx, http.MethodGet, strURL, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (c Client) TunnelLimit(ctx context.Context, brokerID bson.ObjectID, req *mbreq.TunnelLimit) error {
	reqURL := muxproto.ManagerToBrokerURL(brokerID.Hex(), "/api/v1/tunnel/limit")
	strURL := reqURL.String()

	return c.base.SendJSON(ctx, http.MethodPost, strURL, req, nil)
}
