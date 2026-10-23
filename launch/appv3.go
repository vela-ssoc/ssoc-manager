package launch

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/NVIDIA/gontainer/v2"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-common-mb/profile"
	"github.com/vela-ssoc/ssoc-manager/appv3/restapi"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/vela-ssoc/ssoc-manager/component"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxcentral"
	"gorm.io/gorm"
)

func runV3(ctx context.Context, cfg *profile.ManagerConfig, route shipx.RouteGroup, hub linkhub.Huber, hubv3 muxproto.Huber, cli *http.Client, db *gorm.DB, log *slog.Logger) error {
	gopts := []gontainer.Option{
		gontainer.NewService(ctx),
		gontainer.NewService(cfg),
		gontainer.NewService(hub),
		gontainer.NewService(hubv3),
		gontainer.NewService(cli),
		gontainer.NewService(cli.Transport),
		gontainer.NewService(db),
		gontainer.NewService(log),
		//gontainer.NewService(dbconn.MongoDBConfig{URI: cfg.Database.URI}),
		//
		//gontainer.NewFactory(dbconn.OpenMongoDB),
		gontainer.NewFactory(repository.NewBaseDB),
		gontainer.NewFactory(polydb.NewSource),
		gontainer.NewEntrypoint(component.CreateIndex),

		gontainer.NewService(muxcentral.Options{
			Validate:   route.Ship.Validator.Validate,
			Handler:    http.NewServeMux(),
			Logger:     log,
			PerTimeout: 10 * time.Second,
		}),
		gontainer.NewFactory(muxcentral.NewServer),

		gontainer.NewFactory(service.NewBroker),
		gontainer.NewFactory(service.NewBrokerConnectHistory),
		gontainer.NewFactory(service.NewBrokerStream),
		gontainer.NewFactory(service.NewReverseProxy),
		gontainer.NewFactory(service.NewZipArchive),
		gontainer.NewFactory(service.NewZipFile),
		gontainer.NewFactory(restapi.NewBroker),
		gontainer.NewFactory(restapi.NewBrokerConnectHistory),
		gontainer.NewFactory(restapi.NewBrokerReverseProxy),
		gontainer.NewFactory(restapi.NewBrokerStream),
		gontainer.NewFactory(restapi.NewConfig),
		gontainer.NewFactory(restapi.NewReverseProxy),
		gontainer.NewFactory(restapi.NewZipArchive),
		gontainer.NewFactory(restapi.NewTunnel),
		gontainer.NewFactory(restapi.NewZipFile),

		gontainer.NewEntrypoint(route.RegisterHTTP),
	}
	if err := gontainer.Run(gopts...); err != nil {
		return err
	}

	return nil
}
