package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type MinionBinaryService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.MinionBin)
	Deprecate(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	Create(ctx context.Context, req *param.NodeBinaryCreate) error
}

func MinionBinary(gfs gridfs.FS) MinionBinaryService {
	return &minionBinaryService{
		gfs: gfs,
	}
}

type minionBinaryService struct {
	gfs gridfs.FS
}

func (biz *minionBinaryService) Page(ctx context.Context, page param.Pager) (int64, []*model.MinionBin) {
	tbl := query.MinionBin
	dao := tbl.WithContext(ctx)
	if kw := page.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw)).
			Or(tbl.Goos.Like(kw)).
			Or(tbl.Semver.Like(kw))
	}
	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	dats, _ := dao.Order(tbl.Weight.Desc()).
		Order(tbl.UpdatedAt.Desc()).
		Scopes(page.Scope(count)).
		Find()

	return count, dats
}

func (biz *minionBinaryService) Deprecate(ctx context.Context, id int64) error {
	tbl := query.MinionBin
	bin, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if bin.Deprecated {
		return errcode.ErrDeprecated
	}
	if _, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Where(tbl.Deprecated.Is(false)).
		UpdateColumnSimple(tbl.Deprecated.Value(true)); err != nil {
		return err
	}

	return err
}

func (biz *minionBinaryService) Delete(ctx context.Context, id int64) error {
	// 先查询数据
	tbl := query.MinionBin
	dao := tbl.WithContext(ctx).Where(tbl.ID.Eq(id))
	old, err := dao.First()
	if err != nil {
		return err
	}

	_, err = dao.Delete()
	if err == nil {
		_ = biz.gfs.Remove(old.FileID)
	}

	return err
}

func (biz *minionBinaryService) Create(ctx context.Context, req *param.NodeBinaryCreate) error {
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	semver := string(req.Semver)
	tbl := query.MinionBin
	// 检查该发行版是否已经存在
	count, err := tbl.WithContext(ctx).
		Where(tbl.Goos.Eq(req.Goos), tbl.Arch.Eq(req.Arch), tbl.Semver.Eq(semver)).
		Count()
	if count != 0 {
		return errcode.ErrAlreadyExist
	}

	// 将文件保存到数据库
	inf, err := biz.gfs.Write(file, req.Name)
	if err != nil {
		return err
	}

	version := req.Semver.Int64()
	dat := &model.MinionBin{
		FileID:    inf.ID(),
		Goos:      req.Goos,
		Arch:      req.Arch,
		Name:      req.Name,
		Size:      inf.Size(),
		Hash:      inf.MD5(),
		Semver:    req.Semver,
		Changelog: req.Changelog,
		Weight:    version,
	}
	err = tbl.WithContext(ctx).Create(dat)
	if err != nil {
		_ = biz.gfs.Remove(inf.ID())
	}

	return err
}
