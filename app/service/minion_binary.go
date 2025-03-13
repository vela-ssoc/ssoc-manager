package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mresponse"
	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
)

func NewMinionBinary(qry *query.Query, pusher push.Pusher, gfs gridfs.FS) *MinionBinary {
	return &MinionBinary{
		qry:    qry,
		pusher: pusher,
		gfs:    gfs,
	}
}

type MinionBinary struct {
	qry    *query.Query
	pusher push.Pusher
	gfs    gridfs.FS
}

func (biz *MinionBinary) Page1(ctx context.Context, page param.Pager) (int64, []*model.MinionBin) {
	tbl := biz.qry.MinionBin
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

func (biz *MinionBinary) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionBin) {
	tbl := biz.qry.MinionBin
	db := tbl.WithContext(ctx).
		Order(tbl.Weight.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if err := db.Count(&count).Error; err != nil || count == 0 {
		return 0, nil
	}

	ret := make([]*model.MinionBin, 0, page.Size())
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}

func (biz *MinionBinary) Deprecate(ctx context.Context, id int64) error {
	tbl := biz.qry.MinionBin
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

func (biz *MinionBinary) Delete(ctx context.Context, id int64) error {
	// 先查询数据
	tbl := biz.qry.MinionBin
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

func (biz *MinionBinary) Create(ctx context.Context, req *param.NodeBinaryCreate) error {
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	semver := string(req.Semver)
	tbl := biz.qry.MinionBin
	// 检查该发行版是否已经存在
	count, err := tbl.WithContext(ctx).
		Where(tbl.Goos.Eq(req.Goos), tbl.Arch.Eq(req.Arch), tbl.Semver.Eq(semver), tbl.Customized.Eq(req.Customized)).
		Count()
	if count != 0 {
		return errcode.ErrAlreadyExist
	}

	// 将文件保存到数据库
	inf, err := biz.gfs.Write(file, req.Name)
	if err != nil {
		return err
	}

	version := req.Semver.Uint64()
	dat := &model.MinionBin{
		FileID:     inf.ID(),
		Goos:       req.Goos,
		Arch:       req.Arch,
		Name:       req.Name,
		Customized: req.Customized,
		Unstable:   req.Unstable,
		Caution:    req.Caution,
		Ability:    req.Ability,
		Size:       inf.Size(),
		Hash:       inf.MD5(),
		Semver:     req.Semver,
		Changelog:  req.Changelog,
		Weight:     version,
	}
	err = tbl.WithContext(ctx).Create(dat)
	if err != nil {
		_ = biz.gfs.Remove(inf.ID())
	}

	return err
}

func (biz *MinionBinary) Release(ctx context.Context, id int64) error {
	tbl := biz.qry.MinionBin
	bin, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if bin.Deprecated {
		return errcode.ErrDeprecated
	}
	if bin.Unstable {
		return errcode.ErrReleaseUnstable
	}

	go biz.sendRelease(bin)

	return nil
}

func (biz *MinionBinary) Classify(ctx context.Context) ([]*param.MinionBinaryClassify, error) {
	// 查询定制化版本的种类
	tbl := biz.qry.MinionBin

	bins, err := tbl.WithContext(ctx).
		Select(tbl.Goos, tbl.Arch, tbl.Customized).
		Group(tbl.Goos, tbl.Arch, tbl.Customized).
		Order(tbl.Customized, tbl.Goos, tbl.Arch.Desc()).Find()
	if err != nil {
		return nil, err
	}

	ret := make([]*param.MinionBinaryClassify, 0, 16)
	index := make(map[string]*param.MinionBinaryClassify, 16)
	for _, bin := range bins {
		custom := bin.Customized
		classify := index[custom]
		if classify == nil {
			classify = &param.MinionBinaryClassify{
				Structures: make([]*param.MinionBinaryStructure, 0, 4),
				Customized: custom,
			}
			index[custom] = classify
			ret = append(ret, classify)
		}
		st := &param.MinionBinaryStructure{Goos: bin.Goos, Arch: bin.Arch}
		classify.Structures = append(classify.Structures, st)
	}

	return ret, nil
}

func (biz *MinionBinary) Download(ctx context.Context, id int64) (gridfs.File, error) {
	tbl := biz.qry.MinionBin
	bin, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}

	return biz.gfs.OpenID(bin.FileID)
}

func (biz *MinionBinary) Update(ctx context.Context, req *param.MinionBinaryUpdate) error {
	tbl := biz.qry.MinionBin
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		UpdateSimple(
			tbl.Ability.Value(req.Ability),
			tbl.Caution.Value(req.Caution),
			tbl.Changelog.Value(req.Changelog),
		)

	return err
}

func (biz *MinionBinary) Supports() mresponse.BinarySupports {
	return mresponse.DefaultBinarySupports()
}

func (biz *MinionBinary) sendRelease(bin *model.MinionBin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	semver := string(bin.Semver)
	deleted := uint8(model.MSDelete)
	tbl := biz.qry.Minion
	dao := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.BrokerID).
		Where(
			tbl.Goos.Eq(bin.Goos),
			tbl.Arch.Eq(bin.Arch),
			tbl.Customized.Eq(bin.Customized),
			tbl.Edition.Neq(semver),
			tbl.Status.Neq(deleted),
		).Order(tbl.ID).
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
			biz.pusher.Upgrade(ctx, bid, mids, semver, bin.Customized)
		}
	}
}

func (biz *MinionBinary) reduce(minions []*model.Minion) map[int64][]int64 {
	hm := make(map[int64][]int64, 16)
	for _, m := range minions {
		mid, bid := m.ID, m.BrokerID
		hm[bid] = append(hm[bid], mid)
	}
	return hm
}
