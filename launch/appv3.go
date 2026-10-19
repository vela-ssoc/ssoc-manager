package launch

import (
	"context"
	"log/slog"

	"github.com/NVIDIA/gontainer/v2"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-common-mb/dbconn"
	"github.com/vela-ssoc/ssoc-common-mb/profile"
	"github.com/vela-ssoc/ssoc-manager/appv3/restapi"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/vela-ssoc/ssoc-manager/component"
	"gorm.io/gorm"
)

func runV3(ctx context.Context, cfg *profile.ManagerConfig, route shipx.RouteGroup, hub linkhub.Huber, db *gorm.DB, log *slog.Logger) error {
	//valid := validation.New()
	//muxopts := muxcentral.Options{
	//	Validate:   valid.Validate,
	//	Handler:    nil,
	//	Loader:     nil,
	//	Logger:     log,
	//	PerTimeout: 30 * time.Second,
	//}

	gopts := []gontainer.Option{
		gontainer.NewService(ctx),
		gontainer.NewService(cfg),
		gontainer.NewService(hub),
		gontainer.NewService(db),
		gontainer.NewService(log),
		gontainer.NewService(dbconn.MongoDBConfig{URI: cfg.Database.URI}),

		gontainer.NewFactory(dbconn.OpenMongoDB),
		gontainer.NewFactory(repository.NewBaseDB),
		gontainer.NewFactory(polydb.NewSource),
		gontainer.NewEntrypoint(component.CreateIndex),

		gontainer.NewFactory(service.NewBroker),
		gontainer.NewFactory(service.NewReverseProxy),
		gontainer.NewFactory(service.NewZipArchive),
		gontainer.NewFactory(service.NewZipFile),
		gontainer.NewFactory(restapi.NewBroker),
		gontainer.NewFactory(restapi.NewConfig),
		gontainer.NewFactory(restapi.NewReverseProxy),
		gontainer.NewFactory(restapi.NewZipArchive),
		// gontainer.NewFactory(restapi.NewTunnel),
		gontainer.NewFactory(restapi.NewZipFile),

		gontainer.NewEntrypoint(route.RegisterHTTP),
	}
	if err := gontainer.Run(gopts...); err != nil {
		return err
	}

	return nil
}
