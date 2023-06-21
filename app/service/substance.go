package service

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/internal/transact"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/datatypes"
	"gorm.io/gen"
)

type SubstanceService interface {
	Indices(ctx context.Context, idx param.Indexer) []*param.IDName
	Page(ctx context.Context, page param.Pager) (int64, []*param.SubstanceSummary)
	Detail(ctx context.Context, id int64) (*model.Substance, error)
	Create(ctx context.Context, sc *param.SubstanceCreate, userID int64) error
	Update(ctx context.Context, su *param.SubstanceUpdate, userID int64) error
	Delete(ctx context.Context, id int64) error
	Reload(ctx context.Context, mid, sid int64) error
	Resync(ctx context.Context, mid int64) error
	Command(ctx context.Context, mid int64, cmd string) error
}

func Substance(pusher push.Pusher, digest DigestService, sequence SequenceService) SubstanceService {
	return &substanceService{
		pusher:   pusher,
		digest:   digest,
		sequence: sequence,
	}
}

type substanceService struct {
	mutex    sync.Mutex
	pusher   push.Pusher
	digest   DigestService
	sequence SequenceService
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
	var inet string
	tbl := query.Substance
	if mid != 0 {
		// 检查节点
		monTbl := query.Minion
		mon, err := monTbl.WithContext(ctx).
			Select(monTbl.Status, monTbl.BrokerID, monTbl.Inet).
			Where(monTbl.ID.Eq(mid)).
			First()
		if err != nil {
			return errcode.ErrNodeNotExist
		}
		status := mon.Status
		if status != model.MSOffline && status != model.MSOnline {
			return errcode.ErrNodeStatus
		}
		// 私有配置检查配置名是否存在
		if count, err := tbl.WithContext(ctx).
			Where(tbl.Name.Eq(name), tbl.MinionID.Eq(mid)).
			Or(tbl.Name.Eq(name), tbl.MinionID.Eq(0)).
			Count(); count != 0 || err != nil {
			return errcode.FmtErrNameExist.Fmt(name)
		}

		bid = mon.BrokerID
		inet = mon.Inet
	} else { // 公有配置检查名字是否重复
		if count, err := tbl.WithContext(ctx).
			Where(tbl.Name.Eq(name)).
			Count(); err != nil || count != 0 {
			return errcode.FmtErrNameExist.Fmt(name)
		}
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

	if mid != 0 { // 推送
		biz.pusher.TaskSync(ctx, bid, mid, inet)
	}

	return nil
}

func (biz *substanceService) Update(ctx context.Context, su *param.SubstanceUpdate, userID int64) error {
	biz.mutex.Lock()
	defer biz.mutex.Unlock()

	// 1. 查询数据库中原有的数据
	id, version := su.ID, su.Version
	tbl := query.Substance
	sub, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}
	if sub.Version != su.Version {
		return errcode.ErrVersion
	}

	sum := biz.digest.SumMD5(su.Chunk)
	change := sum != sub.Hash

	sub.Hash = sum
	sub.Chunk = su.Chunk
	sub.Icon = su.Icon
	sub.Desc = su.Desc
	sub.UpdatedID = userID
	sub.Version = version + 1

	if _, err = tbl.WithContext(ctx).Where(tbl.Version.Eq(version)).Updates(sub); err != nil || !change {
		return err
	}

	if mid := sub.MinionID; mid != 0 {
		monTbl := query.Minion
		mon, err := monTbl.WithContext(ctx).
			Select(monTbl.ID, monTbl.BrokerID, monTbl.Inet).
			Where(monTbl.ID.Eq(mid)).
			First()
		if err == nil {
			biz.pusher.TaskSync(ctx, mon.BrokerID, mid, mon.Inet)
		}
		return nil
	}

	// 查询所有相关 tag
	// SELECT DISTINCT tag
	// FROM effect
	// WHERE effect_id = ? AND compound = FALSE
	//   OR (compound = TRUE AND effect_id IN (SELECT id FROM compound WHERE JSON_CONTAINS(substances, JSON_ARRAY(?))))
	effTbl := query.Effect
	comTbl := query.Compound
	column := comTbl.Substances.ColumnName().String()

	inSQL := comTbl.WithContext(ctx).Distinct(comTbl.ID).
		Where(gen.Cond(datatypes.JSONArrayQuery(column).Contains(su.ID))...).
		Or(gen.Cond(datatypes.JSONArrayQuery(column).Contains(strconv.FormatInt(su.ID, 10)))...)
	effCtx := effTbl.WithContext(ctx)
	orSQL := effCtx.Where(effTbl.Enable.Is(true)).
		Where(effTbl.Compound.Is(true)).
		Where(effCtx.Columns(effTbl.EffectID).In(inSQL))

	var tags []string
	err = effTbl.WithContext(ctx).
		Distinct(effTbl.Tag).
		Where(effTbl.Enable.Is(true)).
		Where(effTbl.EffectID.Eq(id)).
		Where(effTbl.Compound.Is(false)).
		Or(orSQL).
		Scan(&tags)
	if err != nil || len(tags) == 0 {
		return err
	}

	taskID := biz.sequence.Generate()
	go func() {
		bids, err := transact.EffectTaskTx(ctx, taskID, tags)
		if err == nil && len(bids) != 0 {
			biz.pusher.TaskTable(ctx, bids, taskID)
		}
	}()

	return err
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
	if _, err = subTbl.WithContext(ctx).Delete(dat); err != nil {
		return err
	}

	// 私有配置通知节点
	if mid != 0 {
		monTbl := query.Minion
		mon, err := monTbl.WithContext(ctx).
			Select(monTbl.ID, monTbl.BrokerID, monTbl.Inet).
			Where(monTbl.ID.Eq(mid)).
			First()
		if err == nil {
			biz.pusher.TaskSync(ctx, mon.BrokerID, mid, mon.Inet)
		}
	}

	return nil
}

// Reload 命令 agent 节点重新加载指定配置。
// 该配置必须在该 agent 上发布且已启用，注意要防止越权重启。
func (biz *substanceService) Reload(ctx context.Context, mid, sid int64) error {
	// 检查 minion 节点
	monTbl := query.Minion
	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.ID, monTbl.Inet, monTbl.Status, monTbl.BrokerID).
		Where(monTbl.ID.Eq(mid)).
		First()
	if err != nil {
		return err
	}
	status := mon.Status
	if status != model.MSOnline && status != model.MSOffline {
		return errcode.ErrNodeStatus
	}

	// 1. 查询配置是否存在
	tbl := query.Substance
	sub, err := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.MinionID).
		Where(tbl.ID.Eq(sid)).
		First()
	if err != nil {
		return err
	}
	if did := sub.MinionID; did != 0 && did != mid {
		return errcode.ErrExceedAuthority
	}

	biz.pusher.TaskDiff(ctx, mon.BrokerID, mid, sid, mon.Inet)

	return nil
}

// Resync 重新同步节点上的配置状态
func (biz *substanceService) Resync(ctx context.Context, mid int64) error {
	// 检查 minion 节点
	monTbl := query.Minion
	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.ID, monTbl.Inet, monTbl.Status, monTbl.BrokerID).
		Where(monTbl.ID.Eq(mid)).
		First()
	if err != nil {
		return err
	}
	status := mon.Status
	if status != model.MSOnline && status != model.MSOffline {
		return errcode.ErrNodeStatus
	}

	biz.pusher.TaskSync(ctx, mon.BrokerID, mid, mon.Inet)

	return nil
}

// Command 向指定节点发送指令
func (biz *substanceService) Command(ctx context.Context, mid int64, cmd string) error {
	// 检查 minion 节点
	monTbl := query.Minion
	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.ID, monTbl.Status, monTbl.BrokerID).
		Where(monTbl.ID.Eq(mid)).
		First()
	if err != nil {
		return err
	}
	status := mon.Status
	if status != model.MSOnline && status != model.MSOffline {
		return errcode.ErrNodeStatus
	}

	biz.pusher.Command(ctx, mon.BrokerID, mid, cmd)

	return nil
}
