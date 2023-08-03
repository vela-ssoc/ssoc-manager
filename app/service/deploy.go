package service

import (
	"context"
	"io"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/stegano"
	"github.com/vela-ssoc/vela-common-mb/storage/v2"
	"github.com/vela-ssoc/vela-common-mba/definition"
	"github.com/vela-ssoc/vela-manager/app/internal/modview"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type DeployService interface {
	LAN(ctx context.Context) string
	Script(ctx context.Context, goos string, data *modview.Deploy) (io.Reader, error)
	OpenMinion(ctx context.Context, req *param.DeployMinionDownload) (gridfs.File, error)
}

func Deploy(store storage.Storer, gfs gridfs.FS) DeployService {
	return &deployService{
		store: store,
		gfs:   gfs,
	}
}

type deployService struct {
	store storage.Storer
	gfs   gridfs.FS
}

func (biz *deployService) LAN(ctx context.Context) string {
	addr, _ := biz.store.LocalAddr(ctx)
	return addr
}

func (biz *deployService) OpenMinion(ctx context.Context, req *param.DeployMinionDownload) (gridfs.File, error) {
	// 查询 broker 节点信息
	brkTbl := query.Broker
	brk, err := brkTbl.WithContext(ctx).Where(brkTbl.ID.Eq(req.BrokerID)).First()
	if err != nil {
		return nil, err
	}

	// 查询客户端二进制信息
	version := string(req.Version)
	bin, err := biz.suitableMinion(ctx, req.ID, req.Goos, req.Arch, version)
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
	if file, exx := stegano.AppendStream(inf, hide); exx != nil {
		_ = inf.Close()
		return nil, exx
	} else {
		return file, nil
	}
}

func (biz *deployService) Script(ctx context.Context, goos string, data *modview.Deploy) (io.Reader, error) {
	buf := biz.store.DeployScript(ctx, goos, data)
	return buf, nil
}

func (biz *deployService) suitableMinion(ctx context.Context, id int64, goos, arch, version string) (*model.MinionBin, error) {
	tbl := query.MinionBin
	dao := tbl.WithContext(ctx).
		Where(tbl.Deprecated.Is(false)).
		Order(tbl.Semver.Desc(), tbl.UpdatedAt.Desc())
	if id != 0 {
		return dao.Where(tbl.ID.Eq(id)).First()
	}

	if version != "" {
		return dao.Where(tbl.Goos.Eq(goos), tbl.Arch.Eq(arch), tbl.Semver.Eq(version)).First()
	}

	// 版本号包含 - + 的权重会下降，例如：
	// 0.0.1-debug < 0.0.1
	// 0.0.1+20230720 < 0.0.1
	stmt := dao.Where(tbl.Goos.Eq(goos), tbl.Arch.Eq(arch))
	bin, err := stmt.WithContext(ctx).
		Where(tbl.Semver.NotLike("%-%"), tbl.Semver.NotLike("%+%")).
		First()
	if err == nil {
		return bin, nil
	}

	return stmt.First()
}
