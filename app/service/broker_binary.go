package service

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/param/negotiate"
	"github.com/vela-ssoc/vela-common-mb/storage/v2"
	"github.com/vela-ssoc/vela-common-mba/ciphertext"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/vela-ssoc/vela-manager/param/mresponse"
)

func NewBrokerBinary(qry *query.Query, gfs gridfs.FS, store storage.Storer) *BrokerBinary {
	return &BrokerBinary{
		qry:   qry,
		gfs:   gfs,
		store: store,
	}
}

type BrokerBinary struct {
	qry       *query.Query
	gfs       gridfs.FS
	store     storage.Storer
	uploading atomic.Bool
}

func (biz *BrokerBinary) Page(ctx context.Context, page param.Pager) (int64, []*model.BrokerBin) {
	tbl := biz.qry.BrokerBin
	dao := tbl.WithContext(ctx).Order(tbl.Semver.Desc(), tbl.UpdatedAt.Desc())
	if kw := page.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw)).Or(tbl.Changelog.Like(kw))
	}
	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *BrokerBinary) Create(ctx context.Context, req *param.NodeBinaryCreate) error {
	if !biz.uploading.CompareAndSwap(false, true) {
		return errcode.ErrTaskBusy
	}
	defer biz.uploading.Store(false)

	tbl := biz.qry.BrokerBin

	if count, _ := tbl.WithContext(ctx).
		Where(tbl.Semver.Eq(string(req.Semver)), tbl.Goos.Eq(req.Goos), tbl.Arch.Eq(req.Arch)).
		Count(); count != 0 {
		return errcode.ErrAlreadyExist
	}

	file, err := req.File.Open()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	inf, err := biz.gfs.Write(file, req.Name)
	if err != nil {
		return err
	}

	fid := inf.ID()
	semver := req.Semver
	semverWeight := semver.Uint64()
	bin := &model.BrokerBin{
		Name:         req.Name,
		FileID:       fid,
		Size:         inf.Size(),
		Hash:         inf.MD5(),
		Goos:         req.Goos,
		Arch:         req.Arch,
		Semver:       semver,
		SemverWeight: semverWeight,
		Changelog:    req.Changelog,
	}
	if err = tbl.WithContext(ctx).Create(bin); err != nil {
		_ = biz.gfs.Remove(fid)
		return err
	}

	return nil
}

func (biz *BrokerBinary) Delete(ctx context.Context, id int64) error {
	tbl := biz.qry.BrokerBin
	bin, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if err = biz.gfs.Remove(bin.FileID); err != nil {
		return err
	}
	_, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()

	return err
}

func (biz *BrokerBinary) Open(ctx context.Context, bid, fid int64, addr net.Addr, host string) (gridfs.File, error) {
	tbl := biz.qry.Broker
	brk, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(bid)).First()
	if err != nil {
		return nil, err
	}

	binTbl := biz.qry.BrokerBin
	bin, err := binTbl.WithContext(ctx).Where(binTbl.ID.Eq(fid)).First()
	if err != nil {
		return nil, err
	}

	servers := make(netutil.Addresses, 0, 2)
	if dest, err := biz.store.LocalAddr(ctx); err == nil && dest != "" {
		servers = append(servers, &netutil.Address{Addr: dest, Name: host})
	}
	if addr != nil {
		servers = append(servers, &netutil.Address{Addr: addr.String(), Name: host})
	}
	hide := &negotiate.Hide{
		ID:      bid,
		Secret:  brk.Secret,
		Semver:  string(bin.Semver),
		Servers: servers,
	}
	enc, exx := ciphertext.EncryptPayload(hide)
	if exx != nil {
		return nil, exx
	}

	gf, err := biz.gfs.OpenID(bin.FileID)
	if err != nil {
		return nil, err
	}

	file := gridfs.Merge(gf, enc)

	return file, nil
}

func (biz *BrokerBinary) Latest(ctx context.Context, goos, arch string) *model.BrokerBin {
	tbl := biz.qry.BrokerBin
	bin, _ := tbl.WithContext(ctx).
		Where(tbl.Goos.Eq(goos), tbl.Arch.Eq(arch)).
		Order(tbl.SemverWeight.Desc()).
		First()

	return bin
}

func (biz *BrokerBinary) Supports() mresponse.BinarySupports {
	return mresponse.BinarySupports{
		{
			Name:  "Linux",
			Value: "linux",
			Architectures: mresponse.NameValues{
				{Name: "x64", Value: "amd64"},
				{Name: "x86", Value: "386"},
				{Name: "ARM64", Value: "arm64"},
			},
		},
	}
}
