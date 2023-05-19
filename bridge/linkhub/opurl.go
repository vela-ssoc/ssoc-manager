package linkhub

import (
	"net/url"
	"strconv"
)

const (
	HeaderXNodeID       = "X-Node-Id"
	HeaderXNodeIdentify = "X-Node-Identify"
	QueryNodeKey        = "node"
)

type URLer interface {
	setHost(id int64) URLer
	URL() *url.URL
}

var OpMinionSubstanceEvent any = opURL{path: "/api/v1/substance/event"}

type opURL struct {
	bid  int64
	path string
}

func (op opURL) setHost(bid int64) URLer {
	op.bid = bid
	return op
}

func (op opURL) URL() *url.URL {
	host := strconv.FormatInt(op.bid, 10)
	return &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   op.path,
	}
}
