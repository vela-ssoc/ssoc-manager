package service

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/datatypes"
)

type SubstanceService interface {
	Indices(ctx context.Context, idx param.Indexer) []*param.IDName
	Page(ctx context.Context, page param.Pager) (int64, []*param.SubstanceSummary)
	Detail(ctx context.Context, id int64) (*model.Substance, error)
	Create(ctx context.Context, sc *param.SubstanceCreate, userID int64) error
	Update(ctx context.Context, su *param.SubstanceUpdate, userID int64) error
}

func Substance(hub linkhub.Huber, digest DigestService) SubstanceService {
	return &substanceService{
		hub:    hub,
		digest: digest,
	}
}

type substanceService struct {
	hub    linkhub.Huber
	digest DigestService
}

func (biz *substanceService) Indices(ctx context.Context, idx param.Indexer) []*param.IDName {
	tbl := query.Substance
	dao := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Name).
		Where(tbl.MinionID.Eq(0))
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw))
	}

	var dats []*param.IDName
	_ = dao.Order(tbl.ID).Scan(&dats)

	return dats
}

func (biz *substanceService) Page(ctx context.Context, page param.Pager) (int64, []*param.SubstanceSummary) {
	tbl := query.Substance
	dao := tbl.WithContext(ctx).
		Where(tbl.MinionID.Eq(0)) // minion_id = 0 就是公有配置
	if kw := page.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw)).
			Or(tbl.Desc.Like(kw))
	}
	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}
	subs, err := dao.Order(tbl.ID).Scopes(page.Scope(count)).Find()
	size := len(subs)
	if err != nil || size == 0 {
		return 0, nil
	}

	dats := make([]*param.SubstanceSummary, 0, size)
	for _, sub := range subs {
		ss := &param.SubstanceSummary{
			ID:        sub.ID,
			Name:      sub.Name,
			Icon:      sub.Icon,
			Hash:      sub.Hash,
			Desc:      sub.Desc,
			Links:     []string{},
			Version:   sub.Version,
			CreatedAt: sub.CreatedAt,
			UpdatedAt: sub.UpdatedAt,
		}
		dats = append(dats, ss)
	}

	return count, dats
}

func (biz *substanceService) Detail(ctx context.Context, id int64) (*model.Substance, error) {
	tbl := query.Substance
	dat, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if dat != nil && dat.Links == nil {
		dat.Links = []string{}
	}

	return dat, err
}

func (biz *substanceService) Create(ctx context.Context, sc *param.SubstanceCreate, userID int64) error {
	now := time.Now()
	name, mid := sc.Name, sc.MinionID

	var bid int64
	if mid != 0 {
		// 检查节点
		tbl := query.Minion
		mon, err := tbl.WithContext(ctx).
			Select(tbl.Status, tbl.BrokerID).
			Where(tbl.ID.Eq(mid)).
			First()
		if err != nil {
			return errcode.ErrNodeNotExist
		}
		status := mon.Status
		if status != model.MSOffline && status != model.MSOnline {
			return errcode.ErrNodeStatus
		}
		bid = mon.BrokerID
	}

	// 检查 name 是否已经存在
	tbl := query.Substance
	if count, err := tbl.WithContext(ctx).
		Where(tbl.Name.Eq(name)).
		Count(); err != nil || count != 0 {
		return errcode.FmtErrNameExist.Fmt(name)
	}

	// 计算 hash
	sum := biz.digest.SumMD5(sc.Chunk)
	dat := &model.Substance{
		Name:      name,
		Icon:      sc.Icon,
		Hash:      sum,
		Desc:      sc.Desc,
		Chunk:     sc.Chunk,
		Links:     []string{},
		MinionID:  mid,
		CreatedID: userID,
		UpdatedID: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := tbl.WithContext(ctx).Create(dat); err != nil {
		return err
	}

	if bid != 0 { // 推送
		// TODO biz.hub.Unicast()
	}

	return nil
}

func (biz *substanceService) Update(ctx context.Context, su *param.SubstanceUpdate, userID int64) error {
	// if st.db.Model(new(model.Compound)).Where("? MEMBER OF (substances)", r.ID).Count(&count); count != 0 {
	//		return errno.ErrHasCompound
	//	}
	comTbl := query.Compound
	column := comTbl.Substances.ColumnName().String()
	var dats []*model.Compound
	ret := comTbl.WithContext(ctx).UnderlyingDB().
		Where(datatypes.JSONArrayQuery(column).Contains(su.ID)).
		Or(datatypes.JSONArrayQuery(column).Contains(strconv.FormatInt(su.ID, 10))).
		Find(&dats)
	log.Println(ret)

	return nil
}

func (biz *substanceService) Delete(ctx context.Context, id int64) error {
	// 查询数据
	subTbl := query.Substance
	dat, err := subTbl.WithContext(ctx).
		Select(subTbl.ID, subTbl.MinionID, subTbl.Name).
		Where(subTbl.ID.Eq(id)).
		First()
	if err != nil {
		return errcode.ErrSubstanceNotExist
	}

	mid := dat.MinionID
	if mid == 0 { // 公有配置删除前检查
		// 1. 公有配置发布后不能被删除
		var count int64
		effTbl := query.Effect
		if count, err = effTbl.WithContext(ctx).
			Where(effTbl.EffectID.Eq(id)).
			Where(effTbl.Compound.Is(true)).
			Count(); err != nil || count != 0 {
			return errcode.ErrSubstanceEffected
		}
		// 2. 公有配置组合成服务后不能删除
		comTbl := query.Compound
		column := comTbl.Substances.ColumnName().String()
		comTbl.WithContext(ctx).UnderlyingDB().
			Where(datatypes.JSONArrayQuery(column).Contains(id)).
			Count(&count)
		if count != 0 {
			return errcode.ErrSubstanceCompounded
		}
	}

	// 删除数据
	ret, err := subTbl.WithContext(ctx).Delete(dat)
	if err != nil || ret.RowsAffected != 0 {
		return err
	}

	// 私有配置通知节点
	if mid != 0 {
		// TODO 通知节点更新
	}

	return errcode.ErrDeleteFailed
}
