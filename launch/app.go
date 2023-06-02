package launch

import (
	"context"
	"net/http"
	"time"

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
	srv := &http.Server{
		Addr:              scf.Addr,
		Handler:           a.handler,
		ReadHeaderTimeout: time.Minute,
		ReadTimeout:       time.Hour,
		WriteTimeout:      time.Hour,
	}

	var err error
	cert, pkey := scf.Cert, scf.Pkey
	if cert == "" || pkey == "" {
		err = srv.ListenAndServe()
	} else {
		err = srv.ListenAndServeTLS(cert, pkey)
	}
	ch <- err
}
