package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewCertificate(svc *service.Certificate) *Certificate {
	return &Certificate{svc: svc}
}

type Certificate struct {
	svc *service.Certificate
}

func (crt *Certificate) RegisterHTTP(g shipx.RouteGroup) error {
	g.ExposeV3.Group.Route("/certificates").GET(crt.list)
	//r.Route("/certificate/download").GET(crt.download)
	g.ExposeV3.Group.Route("/certificate/parse").POST(crt.parse)
	g.ExposeV3.Group.Route("/certificate").
		POST(crt.create).
		PUT(crt.update).
		DELETE(crt.delete)

	return nil
}

func (crt *Certificate) list(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, err := crt.svc.List(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (crt *Certificate) create(c *ship.Context) error {
	var param request.CertificateCreate
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return crt.svc.Create(ctx, param)
}

func (crt *Certificate) parse(c *ship.Context) error {
	var param request.CertificateCreate
	if err := c.Bind(&param); err != nil {
		return err
	}

	ret, err := crt.svc.Parse(param.PublicKey, param.PrivateKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (crt *Certificate) update(c *ship.Context) error {
	var param request.CertificateUpdate
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return crt.svc.Update(ctx, param)
}

func (crt *Certificate) delete(c *ship.Context) error {
	var param request.ObjectID
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()
	ids := []bson.ObjectID{param.OID()}

	return crt.svc.Delete(ctx, ids)
}

//func (crt *Certificate) download(c *ship.Context) error {
//	req := new(request.ObjectIDs)
//	if err := c.BindQuery(req); err != nil {
//		return err
//	}
//	ctx := c.Request().Context()
//
//	now := time.Now()
//	name := "certificate-" + strconv.FormatInt(now.Unix(), 10)
//	extension := ".zip"
//	filename := name + extension
//	contentType := mime.TypeByExtension(extension)
//	if contentType == "" {
//		contentType = ship.MIMEOctetStream
//	}
//	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": filename})
//	c.SetRespHeader(ship.HeaderContentType, contentType)
//	c.SetRespHeader(ship.HeaderContentDisposition, disposition)
//	c.WriteHeader(http.StatusOK)
//
//	zipfs := zip.NewWriter(c.ResponseWriter())
//	//goland:noinspection GoUnhandledErrorResult
//	defer zipfs.Close()
//
//	// 这些字符不允许当作文件名。
//	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_",
//		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
//
//	names := make(map[string]struct{}, 16)
//	for dat, exx := range crt.svc.All(ctx, req.OIDs()) {
//		if exx != nil {
//			return exx
//		}
//
//		cname := replacer.Replace(dat.Name)
//		if _, ok := names[cname]; ok {
//			sid := dat.ID.Hex()
//			cname = cname + "-" + sid
//		}
//		names[cname] = struct{}{}
//
//		cw, err := zipfs.Create(cname + ".crt")
//		if err != nil {
//			return err
//		}
//
//		pubKey, priKey := []byte(dat.PublicKey), []byte(dat.PrivateKey)
//		if _, err = cw.Write(pubKey); err != nil {
//			return err
//		}
//		cw, err = zipfs.Create(cname + ".key")
//		if err != nil {
//			return err
//		}
//		_, err = cw.Write(priKey)
//	}
//
//	return nil
//}
