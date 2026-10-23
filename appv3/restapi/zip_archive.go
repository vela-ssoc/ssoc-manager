package restapi

import (
	"mime"
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/yekazip"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
	"github.com/yeka/zip"
)

type ZipArchive struct {
	svc *service.ZipArchive
}

func NewZipArchive(svc *service.ZipArchive) *ZipArchive {
	return &ZipArchive{svc: svc}
}

func (za *ZipArchive) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/zip-archives").GET(za.page)
	g.ExposeV3.Group.Route("/zip-archive").POST(za.create)
	g.ExposeV3.Group.Route("/zip-archive").PUT(za.update)
	g.ExposeV3.Group.Route("/zip-archive/info").GET(za.info)
	g.ExposeV3.Group.Route("/zip-archive/download").GET(za.download)

	return nil
}

func (za *ZipArchive) page(c *ship.Context) error {
	var param request.PageSizeKeyword
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := za.svc.Page(ctx, param)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (za *ZipArchive) create(c *ship.Context) error {
	var param request.ZipArchiveCreate
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return za.svc.Create(ctx, param)
}

func (za *ZipArchive) update(c *ship.Context) error {
	var param request.ZipArchiveUpdate
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return za.svc.Update(ctx, param)
}

func (za *ZipArchive) info(c *ship.Context) error {
	var param request.Name
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	archive, err := za.svc.Get(ctx, param.Name)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, archive)
}

func (za *ZipArchive) download(c *ship.Context) error {
	var param request.Name
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	archive, zfs, err := za.svc.Open(ctx, param.Name)
	if err != nil {
		return err
	}

	zw := zip.NewWriter(c)
	defer zw.Close()

	params := map[string]string{"filename": archive.Name + ".zip"}
	mediaType := mime.FormatMediaType("attachment", params)
	c.SetRespHeader(ship.HeaderContentDisposition, mediaType)
	c.SetRespHeader(ship.HeaderContentType, "application/zip")
	method := zip.EncryptionMethod(archive.Method)

	return yekazip.AddFS(zw, zfs, method, archive.Password)
}
