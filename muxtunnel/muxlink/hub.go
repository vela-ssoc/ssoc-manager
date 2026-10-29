package muxlink

import (
	"sync"

	"github.com/vela-ssoc/vela-tunnel/muxconn"
)

const brokerInternalTLD = "broker.internal"

type Huber interface {
	Put(slug string, mux *muxconn.Muxer, data any) Node
	Get(slug string) Node
	Del(slug string) Node
	All() []Node
	TLD() string
}

func BuildFQDN(host string) string {
	return host + "." + brokerInternalTLD
}

type safeMap struct {
	tld   string // 一般固定 broker.internal
	mtx   sync.RWMutex
	nodes map[string]Node
}

func (sm *safeMap) Put(slug string, mux *muxconn.Muxer, data any) Node {
	//TODO implement me
	panic("implement me")
}

func (sm *safeMap) Get(slug string) Node {
	//TODO implement me
	panic("implement me")
}

func (sm *safeMap) Del(slug string) Node {
	//TODO implement me
	panic("implement me")
}

func (sm *safeMap) All() []Node {
	//TODO implement me
	panic("implement me")
}

func (sm *safeMap) TLD() string {
	return sm.tld
}

func NewMap() any {
	return &safeMap{
		tld:   brokerInternalTLD,
		nodes: make(map[string]Node, 32),
	}
}
