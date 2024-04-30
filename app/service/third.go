package service

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
	"github.com/vela-ssoc/vela-common-mb-itai/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type ThirdService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Third)
	Create(ctx context.Context, name, desc, customized string, r io.Reader, userID int64) error
	Update(ctx context.Context, id int64, desc, customized string, r io.Reader, userID int64) error
	Download(ctx context.Context, id int64) (gridfs.File, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, keyword string) []*param.ThirdListItem
}

func Third(pusher push.Pusher, gfs gridfs.FS) ThirdService {
	return &thirdService{
		gfs:    gfs,
		pusher: pusher,
	}
}

type thirdService struct {
	gfs    gridfs.FS
	pusher push.Pusher
}

func (biz *thirdService) List(ctx context.Context, keyword string) []*param.ThirdListItem {
	tbl := query.Third
	dao := tbl.WithContext(ctx).Order(tbl.ID)
	if keyword != "" {
		like := "%" + keyword + "%"
		dao.Where(tbl.Name.Like(like)).
			Or(tbl.Desc.Like(like)).
			Or(tbl.Customized.Like(like))
	}
	thirds, _ := dao.Find()
	if len(thirds) == 0 {
		return []*param.ThirdListItem{}
	}

	index := make(map[string][]*model.Third, 16)
	for _, third := range thirds {
		cust := third.Customized
		index[cust] = append(index[cust], third)
	}

	cTbl := query.ThirdCustomized
	custs, _ := cTbl.WithContext(ctx).Find()

	ret := make([]*param.ThirdListItem, 0, len(custs))
	defs := index[""]
	if len(defs) != 0 {
		ret = append(ret, &param.ThirdListItem{
			ThirdCustomized: model.ThirdCustomized{},
			Records:         defs,
		})
	}

	for _, cust := range custs {
		name := cust.Name
		item := &param.ThirdListItem{
			ThirdCustomized: model.ThirdCustomized{
				ID:        cust.ID,
				Name:      cust.Name,
				Icon:      cust.Icon,
				Remark:    cust.Remark,
				UpdatedAt: cust.UpdatedAt,
				CreatedAt: cust.CreatedAt,
			},
			Records: index[name],
		}
		ret = append(ret, item)
	}

	return ret
}

func (biz *thirdService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Third) {
	tbl := query.Third
	db := tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if err := db.Count(&count).Error; err != nil || count == 0 {
		return 0, nil
	}

	ret := make([]*model.Third, 0, page.Size())
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}

func (biz *thirdService) Create(ctx context.Context, name, desc, customized string, r io.Reader, userID int64) error {
	// 将文件名统一转为小写后再查询名字是否重复，因为再 windows 下
	// 文件名不区分大小写，所以在上传时就严格把关，防止下发的三方文件
	// 在 agent 出现不可预知的错误。
	lower := strings.ToLower(name)
	ext := filepath.Ext(lower)
	tbl := query.Third
	dao := tbl.WithContext(ctx).Where(tbl.Name.Eq(name))
	if ext == ".zip" {
		// 如果是可解压文件，那么解压后的名字也不能重复
		base := lower[:len(lower)-4]
		dao.Or(tbl.Name.Eq(base))
	}
	if count, err := dao.Count(); err != nil || count != 0 {
		return errcode.FmtErrNameExist.Fmt(name)
	}
	if customized != "" {
		cusTbl := query.ThirdCustomized
		count, _ := cusTbl.WithContext(ctx).Where(cusTbl.Name.Eq(customized)).Count()
		if count == 0 {
			return errcode.ErrCustomizedNotExists
		}
	}

	file, err := biz.gfs.Write(r, name)
	if err != nil {
		return err
	}
	mod := &model.Third{
		FileID:     file.ID(),
		Name:       name,
		Hash:       file.MD5(),
		Path:       "-",
		Desc:       desc,
		Size:       file.Size(),
		Extension:  ext,
		Customized: customized,
		CreatedID:  userID,
		UpdatedID:  userID,
	}
	err = tbl.WithContext(ctx).Create(mod)
	if err != nil {
		_ = biz.gfs.Remove(file.ID())
	}

	return err
}

// Download 下载文件
func (biz *thirdService) Download(ctx context.Context, id int64) (gridfs.File, error) {
	tbl := query.Third
	th, err := tbl.WithContext(ctx).
		Select(tbl.FileID).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return nil, err
	}

	return biz.gfs.OpenID(th.FileID)
}

func (biz *thirdService) Update(ctx context.Context, id int64, desc, customized string, r io.Reader, userID int64) error {
	tbl := query.Third
	// 查询原有的数据
	th, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}

	if customized != "" && customized != th.Customized {
		cusTbl := query.ThirdCustomized
		count, _ := cusTbl.WithContext(ctx).Where(cusTbl.Name.Eq(customized)).Count()
		if count == 0 {
			return errcode.ErrCustomizedNotExists
		}
	}

	th.Desc = desc
	th.UpdatedID = userID
	th.Customized = customized
	if r == nil { // 不更新文件内容
		err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Save(th)
		return err
	}

	// 上传新文件
	file, err := biz.gfs.Write(r, th.Name)
	if err != nil {
		return err
	}
	_ = file.Close()

	hash := th.Hash
	fileID := th.FileID
	th.FileID = file.ID()
	th.Hash = file.MD5()
	th.Size = file.Size()

	if err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Save(th); err != nil {
		// 未更新成功删除新上传的文件
		_ = biz.gfs.Remove(file.ID())
		return err
	}
	// 更新更改则删除原来的文件
	_ = biz.gfs.Remove(fileID)

	if hash != file.MD5() {
		biz.pusher.ThirdUpdate(ctx, th.Name)
	}

	return nil
}

func (biz *thirdService) Delete(ctx context.Context, id int64) error {
	tbl := query.Third
	// 查询原有的数据
	th, err := tbl.WithContext(ctx).
		Select(tbl.FileID).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}

	if _, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete(); err != nil {
		return err
	}
	_ = biz.gfs.Remove(th.FileID)

	biz.pusher.ThirdDelete(ctx, th.Name)

	return nil
}
