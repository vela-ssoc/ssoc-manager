package linkhub

import (
	"sync"
)

type ErrorFuture struct {
	bid int64
	err error
}

func (sf ErrorFuture) Error() error    { return sf.err }
func (sf ErrorFuture) BrokerID() int64 { return sf.bid }

type silentTask struct {
	wg   *sync.WaitGroup
	ret  chan<- *ErrorFuture
	hub  *brokerHub
	bid  int64
	path string
	req  any
}

func (st *silentTask) Run() {
	defer st.wg.Done()
	err := st.hub.silentJSON(st.bid, st.path, st.req)
	fut := &ErrorFuture{bid: st.bid, err: err}
	st.ret <- fut
}

type resultTask struct {
	wg    *sync.WaitGroup
	huber *brokerHub
	id    int64
	path  string
	req   any
	resp  any
	err   error
}

func (rt *resultTask) Wait() error {
	rt.wg.Wait()
	return rt.err
}

func (rt *resultTask) Run() {
	defer rt.wg.Done()
	rt.err = rt.huber.sendJSON(rt.id, rt.path, rt.req, rt.resp)
}
