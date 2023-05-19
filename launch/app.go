package launch

import (
	"context"
	"net/http"

	"github.com/vela-ssoc/vela-manager/infra/config"
)

type application struct {
	cfg     config.Config
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
	}

	return err
}

func (a *application) listen(ch chan<- error) {
	scf := a.cfg.Server
	cert := scf.Cert
	pkey := scf.Pkey
	srv := &http.Server{
		Addr:    scf.Addr,
		Handler: a.handler,
	}

	var err error
	if cert == "" || pkey == "" {
		err = srv.ListenAndServe()
	} else {
		err = srv.ListenAndServeTLS(cert, pkey)
	}
	ch <- err
}
