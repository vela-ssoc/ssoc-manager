package brkclient

import "github.com/vela-ssoc/ssoc-proto/muxtool"

type Client struct {
	base muxtool.Client
}

func NewClient(base muxtool.Client) Client {
	return Client{base: base}
}

func (c Client) Base() muxtool.Client {
	return c.base
}
