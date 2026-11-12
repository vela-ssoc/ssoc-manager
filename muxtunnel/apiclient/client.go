package apiclient

import (
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/httpclient"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
)

type Client struct {
	hub muxproto.Huber
	cli httpclient.Client
	log *slog.Logger
}

func NewClient(hub muxproto.Huber, cli httpclient.Client, log *slog.Logger) Client {
	return Client{
		hub: hub,
		cli: cli,
		log: log,
	}
}

func (c Client) Forget() ClientForget {
	return ClientForget{
		hub: c.hub,
		cli: c.cli,
		log: c.log,
	}
}
