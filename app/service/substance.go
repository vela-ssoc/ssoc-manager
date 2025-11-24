package service

import (
	"context"
	"sync"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"gorm.io/gen"
)

func NewSubstance(qry *query.Query, pusher push.Pusher, digest DigestService, task *SubstanceTask) *Substance {
	return &Substance{
		qry:    qry,
		pusher: pusher,
		digest: digest,
		task:   task,
	}
}

type Substance struct {
	qry    *query.Query
	mutex  sync.Mutex
	pusher push.Pusher
	digest DigestService
	task   *SubstanceTask
}

func (sst *Substance) Indices(ctx context.Context, idx param.Indexer) request.IDNames {
	tbl := sst.qry.Substance
	dao := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Name).
		Where(tbl.MinionID.Eq(0))
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw))
	}

	var dats request.IDNames
	_ = dao.Order(tbl.ID).Scan(&dats)

	return dats
}

func (sst *Substance) Page(ctx context.Context, page param.Pager) (int64, []*param.SubstanceSummary) {
	tbl := sst.qry.Substance
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
	subs, err := dao.Order(tbl.Priority.Desc(), tbl.ID).Scopes(page.Scope(count)).Find()
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

func (sst *Substance) Detail(ctx context.Context, id int64) (*model.Substance, error) {
	tbl := sst.qry.Substance
	dat, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if dat != nil && dat.Links == nil {
		dat.Links = []string{}
	}

	return dat, err
}

func (sst *Substance) Create(ctx context.Context, sc *param.SubstanceCreate, userID int64) error {
	now := time.Now()
	name, mid := sc.Name, sc.MinionID

	var bid int64
	tbl := sst.qry.Substance
	if mid != 0 {
		// 检查节点
		monTbl := sst.qry.Minion
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
	} else { // 公有配置检查名字是否重复
		if count, err := tbl.WithContext(ctx).
			Where(tbl.Name.Eq(name)).
			Count(); err != nil || count != 0 {
			return errcode.FmtErrNameExist.Fmt(name)
		}
	}

	// 计算 hash
	sum := sst.digest.SumMD5(sc.Chunk)
	dat := &model.Substance{
		Name:      name,
		Icon:      sc.Icon,
		Hash:      sum,
		Desc:      sc.Desc,
		Chunk:     sc.Chunk,
		Links:     []string{},
		MinionID:  mid,
		Priority:  sc.Priority,
		CreatedID: userID,
		UpdatedID: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := tbl.WithContext(ctx).Create(dat); err != nil {
		return err
	}

	if mid != 0 { // 推送
		sst.pusher.TaskSync(ctx, bid, []int64{mid})
	}

	return nil
}

func (sst *Substance) Update(ctx context.Context, su *param.SubstanceUpdate, userID int64) (int64, error) {
	sst.mutex.Lock()
	defer sst.mutex.Unlock()

	// 1. 查询数据库中原有的数据
	id, version := su.ID, su.Version
	tbl := sst.qry.Substance
	sub, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return 0, err
	}
	if sub.Version != su.Version {
		return 0, errcode.ErrVersion
	}

	sum := sst.digest.SumMD5(su.Chunk)
	change := sum != sub.Hash

	sub.Hash = sum
	sub.Chunk = su.Chunk
	sub.Icon = su.Icon
	sub.Desc = su.Desc
	sub.UpdatedID = userID
	sub.Version = version + 1
	sub.Priority = su.Priority

	if mid := sub.MinionID; mid != 0 {
		if _, err = tbl.WithContext(ctx).
			Where(tbl.ID.Eq(su.ID), tbl.Version.Eq(version)).
			Updates(sub); err != nil || !change {
			return 0, err
		}
		monTbl := sst.qry.Minion
		mon, err := monTbl.WithContext(ctx).
			Select(monTbl.ID, monTbl.BrokerID, monTbl.Inet).
			Where(monTbl.ID.Eq(mid)).
			First()
		if err == nil {
			sst.pusher.TaskSync(ctx, mon.BrokerID, []int64{mid})
		}
		return 0, nil
	}

	if err = sst.task.BusyError(ctx); err != nil {
		return 0, err
	}

	if _, err = tbl.WithContext(ctx).
		Where(tbl.Version.Eq(version), tbl.ID.Eq(id)).
		Updates(sub); err != nil || !change {
		return 0, err
	}

	effTbl := sst.qry.Effect
	var tags []string
	err = effTbl.WithContext(ctx).
		Distinct(effTbl.Tag).
		Where(effTbl.Enable.Is(true)).
		Where(effTbl.EffectID.Eq(id)).
		Scan(&tags)
	if err != nil || len(tags) == 0 {
		return 0, err
	}

	return sst.task.AsyncTags(ctx, tags)
}

func (sst *Substance) Delete(ctx context.Context, id int64) error {
	// 查询数据
	subTbl := sst.qry.Substance
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
		effTbl := sst.qry.Effect
		if count, err = effTbl.WithContext(ctx).
			Where(effTbl.EffectID.Eq(id)).
			Count(); err != nil || count != 0 {
			return errcode.ErrSubstanceEffected
		}
	}

	// 删除数据
	if _, err = subTbl.WithContext(ctx).Delete(dat); err != nil {
		return err
	}

	// 私有配置通知节点
	if mid != 0 {
		monTbl := sst.qry.Minion
		mon, err := monTbl.WithContext(ctx).
			Select(monTbl.ID, monTbl.BrokerID, monTbl.Inet).
			Where(monTbl.ID.Eq(mid)).
			First()
		if err == nil {
			sst.pusher.TaskSync(ctx, mon.BrokerID, []int64{mid})
		}
	}

	return nil
}

// Reload 命令 agent 节点重新加载指定配置。
// 该配置必须在该 agent 上发布且已启用，注意要防止越权重启。
func (sst *Substance) Reload(ctx context.Context, mid, sid int64) error {
	// 检查 minion 节点
	monTbl := sst.qry.Minion
	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.ID, monTbl.Inet, monTbl.Status, monTbl.BrokerID).
		Where(monTbl.ID.Eq(mid)).
		First()
	if err != nil {
		return err
	}
	status := mon.Status
	if status != model.MSOnline {
		return errcode.ErrNodeStatus
	}

	// 1. 查询配置是否存在
	tbl := sst.qry.Substance
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
	// TODO 检查是否排除

	sst.pusher.TaskDiff(ctx, mon.BrokerID, mid, sid, mon.Inet)

	return nil
}

// Resync 重新同步节点上的配置状态
func (sst *Substance) Resync(ctx context.Context, mid int64) error {
	// 先清除现有上报的配置
	taskTbl := sst.qry.MinionTask
	taskDao := taskTbl.WithContext(ctx)
	_, _ = taskDao.Where(taskTbl.MinionID.Eq(mid)).Delete()

	// 检查 minion 节点
	monTbl := sst.qry.Minion
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

	sst.pusher.TaskSync(ctx, mon.BrokerID, []int64{mid})

	return nil
}

// Command 向指定节点发送指令
func (sst *Substance) Command(ctx context.Context, mid int64, cmd string) error {
	// 检查 minion 节点
	monTbl := sst.qry.Minion
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

	sst.pusher.Command(ctx, mon.BrokerID, []int64{mid}, cmd)

	return nil
}

// Exclude 将某个节点排除某个配置不下发。
func (sst *Substance) Exclude(ctx context.Context, minionID, subID int64) error {
	// 查询节点信息
	monTbl := sst.qry.Minion
	monDao := monTbl.WithContext(ctx)
	if cnt, err := monDao.Where(monTbl.ID.Eq(minionID)).Count(); err != nil {
		return err
	} else if cnt == 0 {
		return errcode.ErrNodeNotExist
	}

	// 查询配置信息
	subTbl := sst.qry.Substance
	subDao := subTbl.WithContext(ctx)
	if cnt, err := subDao.Where(subTbl.ID.Eq(subID)).Count(); err != nil {
		return err
	} else if cnt == 0 {
		return errcode.ErrSubstanceNotExist
	}

	// 排除节点
	mod := &model.MinionSubstanceExclude{
		MinionID:    minionID,
		SubstanceID: subID,
	}
	mseTbl := sst.qry.MinionSubstanceExclude
	mseDao := mseTbl.WithContext(ctx)
	if err := mseDao.Create(mod); err != nil {
		return err
	}

	// 通知节点刷新配置
	_ = sst.Resync(ctx, minionID)

	return nil
}

// Unexclude 移出排除列表。
func (sst *Substance) Unexclude(ctx context.Context, minionID, subID int64) error {
	mseTbl := sst.qry.MinionSubstanceExclude
	mseDao := mseTbl.WithContext(ctx)

	wheres := []gen.Condition{mseTbl.MinionID.Eq(minionID), mseTbl.SubstanceID.Eq(subID)}
	ret, err := mseDao.Where(wheres...).Delete()
	if err != nil {
		return err
	} else if ret.RowsAffected == 0 {
		return nil
	}

	// 通知节点刷新配置
	_ = sst.Resync(ctx, minionID)

	return nil
}
