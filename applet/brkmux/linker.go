package brkmux

import (
	"context"
	"net"
	"strconv"
	"sync"

	"github.com/vela-ssoc/vela-common-mba/smux"
)

type Linker interface {
	// OpenStream 开启一个虚拟 TCP 通道。
	OpenStream(brkID int64) (*smux.Stream, error)

	// DialContext 实现网络拨号方法。
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)

	Brokers() []Broker
}

func newLinker() *brokerLink {
	return &brokerLink{
		brks: make(map[string]*brokerConn, 16),
	}
}

type brokerLink struct {
	mutex sync.RWMutex
	brks  map[string]*brokerConn
}

func (blk *brokerLink) OpenStream(brkID int64) (*smux.Stream, error) {
	id := blk.stringifyAddr(brkID)
	blk.mutex.RLock()
	brk := blk.brks[id]
	blk.mutex.RUnlock()
	if brk == nil {
		return nil, net.ErrClosed
	}

	return brk.mux.OpenStream()
}

func (blk *brokerLink) DialContext(_ context.Context, _ string, address string) (net.Conn, error) {
	id := blk.parseAddr(address)
	blk.mutex.RLock()
	brk := blk.brks[id]
	blk.mutex.RUnlock()
	if brk == nil {
		return nil, net.ErrClosed
	}

	stm, err := brk.mux.OpenStream()
	if err != nil {
		return nil, err
	}

	return stm, nil
}

func (blk *brokerLink) Brokers() []Broker {
	brks := make([]Broker, 0, 20)
	blk.mutex.RLock()
	for _, brk := range blk.brks {
		brks = append(brks, brk)
	}
	blk.mutex.RUnlock()

	return brks
}

func (blk *brokerLink) putConn(bc *brokerConn) {
	id := strconv.FormatInt(bc.ident.ID, 10)
	blk.mutex.Lock()
	blk.brks[id] = bc
	blk.mutex.Unlock()
}

func (blk *brokerLink) delConn(id int64) {
	sid := strconv.FormatInt(id, 10)
	blk.mutex.Lock()
	delete(blk.brks, sid)
	blk.mutex.Unlock()
}

func (*brokerLink) parseAddr(addr string) string {
	if host, _, _ := net.SplitHostPort(addr); host != "" {
		return host
	}
	return addr
}

func (*brokerLink) stringifyAddr(brkID int64) string {
	return strconv.FormatInt(brkID, 10)
}
