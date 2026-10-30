package launch

import (
	"context"
	"net/http"
	"time"

	"github.com/NVIDIA/gontainer/v2"
	"github.com/vela-ssoc/ssoc-common-mb/crontab"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

func closeMDB(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cli := db.Client()

	return cli.Disconnect(ctx)
}

func closeRDB(db *gorm.DB) error {
	sdb, err := db.DB()
	if err != nil {
		return err
	}

	return sdb.Close()
}

func startHTTPS(errch chan error, srv *http.Server) {
	errch <- srv.ListenAndServeTLS("", "")
}

func closeHTTP(srv *http.Server) error {
	return srv.Close()
}

func registerRoutes(route shipx.RouteGroup3, handlers gontainer.Multiple[shipx.HTTPRegistrar]) error {
	return route.RegisterRoutes(handlers)
}

func registerV1Routes(route shipx.RouteGroup3, handlers gontainer.Multiple[shipx.V1HTTPRegistrar]) {
	route.RegisterV1Routes(handlers)
}

func cronAddJobs(c *crontab.Table, jobs gontainer.Multiple[crontab.Jober]) error {
	return c.AddJobs(jobs)
}
