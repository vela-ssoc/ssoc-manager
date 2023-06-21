package service

import (
	"bytes"
	"context"
	"io"
	"text/template"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/stegano"
	"github.com/vela-ssoc/vela-common-mba/definition"
	"github.com/vela-ssoc/vela-manager/app/internal/modview"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type DeployService interface {
	LAN(ctx context.Context) string
	Script(ctx context.Context, goos string, data *modview.Deploy) (io.Reader, error)
	OpenMinion(ctx context.Context, req *param.DeployMinionDownload) (gridfs.File, error)
}

func Deploy(store StoreService, gfs gridfs.FS) DeployService {
	return &deployService{
		store: store,
		gfs:   gfs,
	}
}

type deployService struct {
	store StoreService
	gfs   gridfs.FS
}

func (biz *deployService) LAN(ctx context.Context) string {
	const key = "global.local.addr"
	if st, _ := biz.store.FindID(ctx, key); st != nil {
		return string(st.Value)
	}

	return ""
}

func (biz *deployService) OpenMinion(ctx context.Context, req *param.DeployMinionDownload) (gridfs.File, error) {
	// 查询 broker 节点信息
	brkTbl := query.Broker
	brk, err := brkTbl.WithContext(ctx).Where(brkTbl.ID.Eq(req.BrokerID)).First()
	if err != nil {
		return nil, err
	}
	// 查询客户端二进制信息
	tbl := query.MinionBin
	var bin *model.MinionBin
	if req.ID > 0 {
		bin, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(req.ID)).First()
	} else {
		dao := tbl.WithContext(ctx).
			Where(tbl.Goos.Eq(req.Goos), tbl.Arch.Eq(req.Arch)).
			Order(tbl.Weight.Desc(), tbl.UpdatedAt.Desc())
		if req.Version != "" {
			dao.Where(tbl.Semver.Eq(string(req.Version)))
		}
		bin, err = dao.First()
	}
	if err != nil {
		return nil, err
	}

	inf, err := biz.gfs.OpenID(bin.FileID)
	if err != nil {
		return nil, err
	}

	hide := &definition.MinionHide{
		Servername: brk.Servername,
		LAN:        brk.LAN,
		VIP:        brk.VIP,
		Edition:    string(bin.Semver),
		Hash:       inf.MD5(),
		Size:       inf.Size(),
		Tags:       req.Tags,
		DownloadAt: time.Now(),
	}
	file, err := stegano.AppendStream(inf, hide)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (biz *deployService) Script(ctx context.Context, goos string, data *modview.Deploy) (io.Reader, error) {
	id := "global.deploy." + goos + ".tmpl"
	st, err := biz.store.FindID(ctx, id)
	if err != nil {
		return nil, err
	}
	tpl, err := template.New(id).Parse(string(st.Value))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, data); err != nil {
		return nil, err
	}

	return buf, nil
}
