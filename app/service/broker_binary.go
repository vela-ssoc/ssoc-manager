package service

import (
	"context"
	"sync/atomic"

	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type BrokerBinaryService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.BrokerBin)
	Create(ctx context.Context, req *param.NodeBinaryCreate) error
	Delete(ctx context.Context, id int64) error
}

func BrokerBinary(gfs gridfs.FS) BrokerBinaryService {
	return &brokerBinaryService{
		gfs: gfs,
	}
}

type brokerBinaryService struct {
	gfs       gridfs.FS
	uploading atomic.Bool
}

func (biz *brokerBinaryService) Page(ctx context.Context, page param.Pager) (int64, []*model.BrokerBin) {
	tbl := query.BrokerBin
	dao := tbl.WithContext(ctx).Order(tbl.Semver.Desc(), tbl.UpdatedAt.Desc())
	if kw := page.Keyword(); kw != "" {
		dao.Or(tbl.Name.Like(kw), tbl.Changelog.Like(kw))
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
