package service

import (
	"context"
	"log/slog"
	"sync"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/applet/manager/request"
)

func NewAlertServer(qry *query.Query, log *slog.Logger) *AlertServer {
	return &AlertServer{
		qry: qry,
		log: log,
	}
}

type AlertServer struct {
	qry   *query.Query
	log   *slog.Logger
	mutex sync.Mutex
}

func (alt *AlertServer) Find(ctx context.Context) *model.AlertServer {
	tbl := alt.qry.AlertServer
	dat, _ := tbl.WithContext(ctx).First()

	return dat
}

func (alt *AlertServer) Upsert(ctx context.Context, req *request.AlertServerUpsert) error {
	alt.mutex.Lock()
	defer alt.mutex.Unlock()

	return nil
}
