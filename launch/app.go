package launch

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/prereadtls"
	"github.com/vela-ssoc/ssoc-common-mb/profile"
)

type application struct {
	cfg     *profile.ManagerConfig
	handler http.Handler
	parent  context.Context
}

func (a *application) run() error {
	ch := make(chan error, 1)

	go a.listen(ch)

	var err error
	select {
	case err = <-ch:
	case <-a.parent.Done():
		err = a.parent.Err()
	}

	return err
}

func (a *application) listen(ch chan<- error) {
	scf := a.cfg.Server
	lis, err := net.Listen("tcp", scf.Addr)
	if err != nil {
		ch <- err
		return
	}
	//goland:noinspection GoUnhandledErrorResult
	defer lis.Close()

	tcpSrv := &http.Server{
		Handler:           a.handler,
		ReadHeaderTimeout: time.Minute,
		ReadTimeout:       time.Hour,
		WriteTimeout:      time.Hour,
	}

	var tlsFunc func(net.Conn)
	cert, pkey := scf.Cert, scf.Pkey
	if cert != "" && pkey != "" {
		pair, exx := tls.LoadX509KeyPair(cert, pkey)
		if exx != nil {
			ch <- exx
			return
		}

		tcpSrv.Handler = &onlyDeploy{h: tcpSrv.Handler}
		tlsSrv := &http.Server{
			Handler:           a.handler,
			ReadHeaderTimeout: time.Minute,
			ReadTimeout:       time.Hour,
			WriteTimeout:      time.Hour,
			TLSConfig:         &tls.Config{Certificates: []tls.Certificate{pair}},
		}
		tlsFunc = func(conn net.Conn) {
			ln := prereadtls.NewOnceAccept(conn)
			_ = tlsSrv.ServeTLS(ln, "", "")
		}
	}
	tcpFunc := func(conn net.Conn) {
		ln := prereadtls.NewOnceAccept(conn)
		_ = tcpSrv.Serve(ln)
	}

	ch <- prereadtls.Serve(lis, tcpFunc, tlsFunc)
}

type onlyDeploy struct {
	h http.Handler
}

func (od *onlyDeploy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allows := map[string]struct{}{
		"/api/v1/brkbin":                  {},
		"/api/v1/brkbin/":                 {},
		"/api/v1/deploy/minion":           {},
		"/api/v1/deploy/minion/":          {},
		"/api/v1/deploy/minion/download":  {},
		"/api/v1/deploy/minion/download/": {},
	}
	path := r.URL.Path
	if _, allow := allows[path]; allow {
		od.h.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusUpgradeRequired)
	}
}
