package restapi

import (
	"archive/zip"
	"mime"
	"net/http"
	"strconv"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type ZipFile struct {
	svc *service.ZipFile
}

func NewZipFile(svc *service.ZipFile) *ZipFile {
	return &ZipFile{svc: svc}
}

func (zf *ZipFile) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/zip-file/readdir").GET(zf.readdir)   // 获取 zip 下的文件列表（所有文件）。
	g.ExposeV3.Group.Route("/zip-file/content").GET(zf.read)      // 下载文件（预览文件）。
	g.ExposeV3.Group.Route("/zip-file/unzip").POST(zf.unzip)      // 上传一个 zip 文件，读取映射成虚拟 zip 文件。
	g.ExposeV3.Group.Route("/zip-file/mkdir").POST(zf.mkdir)      // 创建目录。
	g.ExposeV3.Group.Route("/zip-file/content").POST(zf.update)   // 更新文件内容（全量更新，不支持部分更新）
	g.ExposeV3.Group.Route("/zip-file/content").DELETE(zf.delete) // 删除 zip 内的某个文件/目录（删除目录时会删除下面的子目录和文件）。

	return nil
}

func (zf *ZipFile) readdir(c *ship.Context) error {
	var param request.Name
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, files, err := zf.svc.List(ctx, param.Name)
	if err != nil {
		return err
	}
	ret := response.NewRecords(files)

	return c.JSON(http.StatusOK, ret)
}

func (zf *ZipFile) unzip(c *ship.Context) error {
	var param request.ZipFileUnzip
	if err := c.Bind(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	size := param.File.Size
	file, err := param.File.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	zr, err := zip.NewReader(file, size)
	if err != nil {
		return err
	}

	return zf.svc.Unzip(ctx, param.ZipArchiveCreate, zr, size)
}

func (zf *ZipFile) mkdir(c *ship.Context) error {
	var param request.ZipFileName
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return zf.svc.Mkdir(ctx, param.Name, param.FullName)
}

func (zf *ZipFile) update(c *ship.Context) error {
	var param request.ZipFileUpdate
	if err := c.Bind(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	file, err := param.File.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	return zf.svc.Update(ctx, param.Name, param.FullName, file)
}

func (zf *ZipFile) read(c *ship.Context) error {
	var param request.ZipFileName
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, file, stm, err := zf.svc.Open(ctx, param.Name, param.FullName)
	if err != nil {
		return err
	}
	defer stm.Close()

	contentLength := strconv.FormatInt(file.FileSize, 10)
	params := map[string]string{"filename": file.FileName}
	mediaType := mime.FormatMediaType("attachment", params)
	c.SetRespHeader(ship.HeaderContentLength, contentLength)
	c.SetRespHeader(ship.HeaderContentDisposition, mediaType)

	return c.Stream(http.StatusOK, file.MIMEType, stm)
}

func (zf *ZipFile) delete(c *ship.Context) error {
	var param request.ZipFileName
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return zf.svc.Delete(ctx, param.Name, param.FullName)
}
