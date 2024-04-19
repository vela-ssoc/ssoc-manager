package service

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
	"github.com/vela-ssoc/vela-common-mb-itai/stegano"
	"github.com/vela-ssoc/vela-common-mb-itai/storage/v2"
	"github.com/vela-ssoc/vela-common-mba/ciphertext"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type BrokerBinaryService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.BrokerBin)
	Create(ctx context.Context, req *param.NodeBinaryCreate) error
	Delete(ctx context.Context, id int64) error
	Open(ctx context.Context, bid, fid int64, eth net.Addr, host string) (gridfs.File, error)
}

func BrokerBinary(gfs gridfs.FS, store storage.Storer) BrokerBinaryService {
	return &brokerBinaryService{
		gfs:   gfs,
		store: store,
	}
}

type brokerBinaryService struct {
	gfs       gridfs.FS
	store     storage.Storer
	uploading atomic.Bool
}

func (biz *brokerBinaryService) Page(ctx context.Context, page param.Pager) (int64, []*model.BrokerBin) {
	tbl := query.BrokerBin
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

func (biz *brokerBinaryService) Create(ctx context.Context, req *param.NodeBinaryCreate) error {
	if !biz.uploading.CompareAndSwap(false, true) {
		return errcode.ErrTaskBusy
	}
	defer biz.uploading.Store(false)

	tbl := query.BrokerBin
	semver := string(req.Semver)
	if count, _ := tbl.WithContext(ctx).
		Where(tbl.Semver.Eq(semver), tbl.Goos.Eq(req.Goos), tbl.Arch.Eq(req.Arch)).
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
	bin := &model.BrokerBin{
		Name:      req.Name,
		FileID:    fid,
		Size:      inf.Size(),
		Hash:      inf.MD5(),
		Goos:      req.Goos,
		Arch:      req.Arch,
		Semver:    req.Semver,
		Changelog: req.Changelog,
	}
	if err = tbl.WithContext(ctx).Create(bin); err != nil {
		_ = biz.gfs.Remove(fid)
		return err
	}

	return nil
}

func (biz *brokerBinaryService) Delete(ctx context.Context, id int64) error {
	tbl := query.BrokerBin
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

func (biz *brokerBinaryService) Open(ctx context.Context, bid, fid int64, addr net.Addr, host string) (gridfs.File, error) {
	tbl := query.Broker
	brk, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(bid)).First()
	if err != nil {
		return nil, err
	}

	binTbl := query.BrokerBin
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
	hide := &stegano.BHide{
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
