package launch

import (
	"context"
	"log/slog"

	"github.com/NVIDIA/gontainer/v2"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-common-mb/dbconn"
	"github.com/vela-ssoc/ssoc-common-mb/profile"
	"github.com/vela-ssoc/ssoc-manager/appv3/manager/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/manager/webapi"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

func runV3(ctx context.Context, cfg *profile.ManagerConfig, v3 *ship.RouteGroupBuilder, log *slog.Logger) error {
	route := shipx.RouteGroup{
		Ship:  v3.Ship(),
		APIV3: v3,
	}

	// 依赖注入
	gopts := []gontainer.Option{
		gontainer.NewService(cfg),
		gontainer.NewService(log),
		gontainer.NewService(dbconn.MongoDBConfig{URI: cfg.Database.URI}),

		gontainer.NewFactory(dbconn.OpenMongoDB),
		gontainer.NewFactory(repository.NewBaseDB),
		gontainer.NewEntrypoint(func(db *repository.BaseDB) error {
			return db.CreateIndex(ctx)
		}),

		gontainer.NewFactory(service.NewReverseProxy),
		gontainer.NewFactory(service.NewZipArchive),
		gontainer.NewFactory(service.NewZipFile),
		gontainer.NewFactory(webapi.NewReverseProxy),
		gontainer.NewFactory(webapi.NewZipArchive),
		gontainer.NewFactory(webapi.NewZipFile),

		gontainer.NewEntrypoint(route.RegisterHTTP),
	}
	if err := gontainer.Run(gopts...); err != nil {
		return err
	}

	return nil
}
