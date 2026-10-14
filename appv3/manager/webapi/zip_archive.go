package webapi

import (
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/manager/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/manager/service"
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

func (za *ZipArchive) RegisterHTTP(g shipx.RouteGroup) error {
	g.APIV3.Route("/zip-archives").GET(za.page)
	g.APIV3.Route("/zip-archive").POST(za.create)
	g.APIV3.Route("/zip-archive").PUT(za.update)
	g.APIV3.Route("/zip-archive/info").GET(za.info)
	g.APIV3.Route("/zip-archive/download").GET(za.download)

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
	enc := zip.EncryptionMethod(archive.Method)

	return za.addFSEncrypted(zw, zfs, enc, archive.Password)
}

// AddFS adds the files from fs.FS to the archive.
// It walks the directory tree starting at the root of the filesystem
// adding each file to the zip using deflate while maintaining the directory structure.
func (*ZipArchive) addFSEncrypted(zw *zip.Writer, fsys fs.FS, method zip.EncryptionMethod, password string) error {
	return fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !d.IsDir() && !info.Mode().IsRegular() {
			return errors.New("zip: cannot add non-regular file")
		}
		h, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		h.Name = name
		if d.IsDir() {
			h.Name += "/"
		}
		h.Method = zip.Deflate

		// 设置加密算法和密码
		if method > 0 && password != "" {
			h.SetEncryptionMethod(method)
			h.SetPassword(password)
		}

		fw, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, err := fsys.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(fw, f)

		return err
	})
}
