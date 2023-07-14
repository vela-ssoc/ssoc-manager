package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type MinionBinaryService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.MinionBin)
	Deprecate(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	Create(ctx context.Context, req *param.NodeBinaryCreate) error
	Release(ctx context.Context, id int64) error
}

func MinionBinary(pusher push.Pusher, gfs gridfs.FS) MinionBinaryService {
	return &minionBinaryService{
		pusher: pusher,
		gfs:    gfs,
	}
}

type minionBinaryService struct {
	pusher push.Pusher
	gfs    gridfs.FS
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

func (biz *minionBinaryService) Release(ctx context.Context, id int64) error {
	tbl := query.MinionBin
	bin, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if bin.Deprecated {
		return errcode.ErrDeprecated
	}

	go biz.sendRelease(bin.Goos, bin.Arch, string(bin.Semver))

	return nil
}

func (biz *minionBinaryService) sendRelease(goos, arch, semver string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	deleted := uint8(model.MSDelete)
	tbl := query.Minion
	dao := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.BrokerID).
		Where(
			tbl.Goos.Eq(goos),
			tbl.Arch.Eq(arch),
			tbl.Edition.Neq(semver),
			tbl.Status.Neq(deleted),
		).
		Order(tbl.ID).
		Limit(200)

	var lastID int64
	for {
		minions, err := dao.Where(tbl.ID.Gt(lastID)).Find()
		size := len(minions)
		if err != nil || size == 0 {
			break
		}
		lastID = minions[size-1].ID
		hm := biz.reduce(minions)
		for bid, mids := range hm {
			biz.pusher.Upgrade(ctx, bid, mids, semver)
		}
	}
}

func (biz *minionBinaryService) reduce(minions []*model.Minion) map[int64][]int64 {
	hm := make(map[int64][]int64, 16)
	for _, m := range minions {
		mid, bid := m.ID, m.BrokerID
		hm[bid] = append(hm[bid], mid)
	}
	return hm
}
