package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/internal/transact"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type CompoundService interface {
	Indices(ctx context.Context, idx param.Indexer) []*param.IDName
	Page(ctx context.Context, page param.Pager) (int64, []*param.CompoundVO)
	Create(ctx context.Context, req *param.CompoundCreate, userID int64) error
	Update(ctx context.Context, req *param.CompoundUpdate, userID int64) error
	Delete(ctx context.Context, id int64) error
}

func Compound(pusher push.Pusher, seq SequenceService) CompoundService {
	return &compoundService{
		pusher: pusher,
		seq:    seq,
	}
}

type compoundService struct {
	pusher push.Pusher
	seq    SequenceService
}

func (biz *compoundService) Indices(ctx context.Context, idx param.Indexer) []*param.IDName {
	tbl := query.Compound
	dao := tbl.WithContext(ctx).Select(tbl.ID, tbl.Name)
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw)).
			Or(tbl.Desc.Like(kw))
	}

	var dats []*param.IDName
	_ = dao.Order(tbl.ID).Scopes(idx.Scope).Scan(&dats)

	return dats
}

func (biz *compoundService) Page(ctx context.Context, page param.Pager) (int64, []*param.CompoundVO) {
	tbl := query.Compound
	dao := tbl.WithContext(ctx)
	if kw := page.Keyword(); kw != "" {
		dao.Or(tbl.Name.Like(kw), tbl.Desc.Like(kw))
	}
	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}
	cms, err := dao.Scopes(page.Scope(count)).Find()
	if err != nil {
		return 0, nil
	}

	var names param.IDNames
	subIDs := model.Compounds(cms).SubstanceIDs()
	subTbl := query.Substance
	_ = subTbl.WithContext(ctx).
		Select(subTbl.ID, subTbl.Name).
		Where(subTbl.ID.In(subIDs...)).
		Scan(&names)

	hm := names.Map()
	ret := make([]*param.CompoundVO, 0, len(cms))

	for _, cm := range cms {
		vo := &param.CompoundVO{
			ID:        cm.ID,
			Name:      cm.Name,
			Desc:      cm.Desc,
			Version:   cm.Version,
			CreatedAt: cm.CreatedAt,
			UpdatedAt: cm.UpdatedAt,
		}
		for _, s := range cm.Substances {
			if idName := hm[s]; idName != nil {
				vo.Substances = append(vo.Substances, idName)
			}
		}
		ret = append(ret, vo)
	}

	return count, ret
}

func (biz *compoundService) Create(ctx context.Context, req *param.CompoundCreate, userID int64) error {
	// 检查文件名是否存在
	tbl := query.Compound
	if count, _ := tbl.WithContext(ctx).
		Where(tbl.Name.Eq(req.Name)).
		Count(); count != 0 {
		return errcode.FmtErrNameExist.Fmt(req.Name)
	}
	// 检查配置是否存在
	subTbl := query.Substance
	count, _ := subTbl.WithContext(ctx).Where(subTbl.ID.In(req.Substances...)).Count()
	if int(count) != len(req.Substances) {
		return errcode.ErrSubstanceNotExist
	}

	dat := &model.Compound{
		Name:       req.Name,
		Desc:       req.Desc,
		Substances: req.Substances,
		CreatedID:  userID,
		UpdatedID:  userID,
	}

	return tbl.WithContext(ctx).Create(dat)
}

func (biz *compoundService) Update(ctx context.Context, req *param.CompoundUpdate, userID int64) error {
	id, version := req.ID, req.Version
	tbl := query.Compound
	old, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if req.Name != old.Name {
		// 检查文件名是否存在
		if count, _ := tbl.WithContext(ctx).
			Where(tbl.Name.Eq(req.Name)).
			Count(); count != 0 {
			return errcode.FmtErrNameExist.Fmt(req.Name)
		}
	}
	// 检查配置是否存在
	subTbl := query.Substance
	count, _ := subTbl.WithContext(ctx).Where(subTbl.ID.In(req.Substances...)).Count()
	if int(count) != len(req.Substances) {
		return errcode.ErrSubstanceNotExist
	}

	old.Name = req.Name
	old.Desc = req.Desc
	old.Substances = req.Substances
	old.UpdatedID = userID
	old.Version = version + 1
	_, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id), tbl.Version.Eq(version)).
		Updates(old)
	if err != nil {
		return err
	}

	// 查询关联的 tags 节点通知更新
	var tags []string
	effTbl := query.Effect
	err = effTbl.WithContext(ctx).
		Where(effTbl.ID.Eq(id), effTbl.Compound.Is(true)).
		Scan(&tags)
	if err != nil || len(tags) == 0 {
		return err
	}

	taskID := biz.seq.Generate()
	brokerIDs, err := transact.EffectTaskTx(ctx, taskID, tags)
	if err != nil {
		return err
	}

	// 推送任务
	biz.pusher.TaskTable(ctx, brokerIDs, taskID)

	return nil
}

func (biz *compoundService) Delete(ctx context.Context, id int64) error {
	tbl := query.Compound
	_, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	// 检查配置是否已经发布，已经发布的配置不允许删除
	effTbl := query.Effect
	count, _ := effTbl.WithContext(ctx).
		Where(effTbl.ID.Eq(id), effTbl.Compound.Is(true)).
		Count()
	if count != 0 {
		return errcode.ErrDeleteFailed
	}
	_, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()

	return err
}
