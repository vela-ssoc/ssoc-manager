package restapi

import (
	"mime"
	"net/http"
	"strconv"

	"github.com/vela-ssoc/vela-manager/applet/manager/request"
	"github.com/vela-ssoc/vela-manager/applet/manager/service"
	"github.com/xgfone/ship/v5"
)

func NewGrid(svc *service.Grid) *Grid {
	return &Grid{svc: svc}
}

type Grid struct {
	svc *service.Grid
}

func (gd *Grid) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/grid/cond").GET(gd.cond)
	r.Route("/grid/files").GET(gd.page)
	r.Route("/grid/upload").PUT(gd.upload)
	r.Route("/grid/download").GET(gd.download)
	return nil
}

func (gd *Grid) cond(c *ship.Context) error {
	dat := gd.svc.Cond()
	return c.JSON(http.StatusOK, dat)
}

func (gd *Grid) page(c *ship.Context) error {
	req := new(request.PageCondition)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	dat, err := gd.svc.Page(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dat)
}

func (gd *Grid) download(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	file, err := gd.svc.Open(ctx, req.ID)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	length := strconv.FormatInt(file.Size(), 10)
	params := map[string]string{"filename": file.Name()}
	disposition := mime.FormatMediaType("attachment", params)
	c.SetRespHeader(ship.HeaderContentDisposition, disposition)
	c.SetRespHeader(ship.HeaderContentLength, length)

	return c.Stream(http.StatusOK, file.MIME(), file)
}

func (gd *Grid) upload(c *ship.Context) error {
	req := new(request.GridUpload)
	if err := c.Bind(req); err != nil {
		return err
	}

	file, err := req.File.Open()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()
	ctx := c.Request().Context()

	return gd.svc.Create(ctx, req.Name, file)
}
