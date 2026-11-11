package apiclient

import "net/http"

type Client struct {
	cli *http.Client
}

func NewClient(cli *http.Client) Client {
	return Client{cli: cli}
}
