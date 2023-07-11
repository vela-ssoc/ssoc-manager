package service

import (
	"context"
	"io"
	"io/fs"
	"strconv"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/stegano"
	"github.com/vela-ssoc/vela-common-mb/storage"
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
	tbl := query.MinionBin
	var bin *model.MinionBin
	if req.ID > 0 {
		bin, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(req.ID)).First()
	} else {
		dao := tbl.WithContext(ctx).
			Where(tbl.Goos.Eq(req.Goos), tbl.Arch.Eq(req.Arch)).
			Order(tbl.Weight.Desc(), tbl.UpdatedAt.Desc())
		if ver := string(req.Version); ver != "" {
			dao.Where(tbl.Semver.Eq(ver))
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
	if file, exx := stegano.AppendStream(inf, hide); exx != nil {
		_ = inf.Close()
		return nil, exx
	} else {
		return file, nil
	}
}

func (biz *deployService) Script(ctx context.Context, goos string, data *modview.Deploy) (io.Reader, error) {
	buf := biz.store.Deploy(ctx, goos, data)
	return buf, nil
}

type oldFile struct {
	size int64
	inf  gridfs.File
	rd   io.Reader
}

func (o *oldFile) Stat() (fs.FileInfo, error) {
	return o.inf.Stat()
}

func (o *oldFile) Read(p []byte) (int, error) {
	return o.rd.Read(p)
}

func (o *oldFile) Close() error {
	return o.inf.Close()
}

func (o *oldFile) Name() string {
	return o.inf.Name()
}

func (o *oldFile) Size() int64 {
	return o.size
}

func (o *oldFile) Mode() fs.FileMode {
	return o.inf.Mode()
}

func (o *oldFile) ModTime() time.Time {
	return o.inf.ModTime()
}

func (o *oldFile) IsDir() bool {
	return o.inf.IsDir()
}

func (o *oldFile) Sys() any {
	return o.inf.Sys()
}

func (o *oldFile) ID() int64 {
	return o.inf.ID()
}

func (o *oldFile) MD5() string {
	return ""
}

func (o *oldFile) ContentType() string {
	return o.inf.ContentType()
}

func (o *oldFile) ContentLength() string {
	return strconv.FormatInt(o.size, 10)
}

func (o *oldFile) Disposition() string {
	return o.inf.Disposition()
}
