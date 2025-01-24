package service

import (
	"log/slog"

	"github.com/vela-ssoc/vela-manager/applet/brkmux"
)

func NewBroker(link brkmux.Linker, log *slog.Logger) *Broker {
	return &Broker{
		link: link,
		log:  log,
	}
}

type Broker struct {
	link brkmux.Linker
	log  *slog.Logger
}

func (bk *Broker) Onlines() []brkmux.Broker {
	return bk.link.Brokers()
}

func (bk *Broker) Kickout(id int64) error {
	for _, brk := range bk.link.Brokers() {
		ident, _ := brk.Info()
		if ident.ID != id {
			continue
		}
		return brk.Close()
	}

	return nil
}
