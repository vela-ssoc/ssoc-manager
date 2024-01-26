package service

import (
	"context"
	"net"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-common-mb/integration/cmdb"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/internal/sheet"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

type MinionService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope, likes []gen.Condition) (int64, []*param.MinionSummary)
	Detail(ctx context.Context, id int64) (*param.MinionDetail, error)
	Drop(ctx context.Context, id int64) error
	Create(ctx context.Context, mc *param.MinionCreate) error
	Delete(ctx context.Context, scope dynsql.Scope, likes []gen.Condition) error
	CSV(ctx context.Context) sheet.CSVStreamer
	Upgrade(ctx context.Context, mid, binID int64) error
	Batch(ctx context.Context, scope dynsql.Scope, likes []gen.Condition, cmd string) error
	Command(ctx context.Context, mid int64, cmd string) error
	Unload(ctx context.Context, mid int64, unload bool) error
	BatchTag(ctx context.Context, scope dynsql.Scope, likes []gen.Condition, creates, deletes []string) error
}

func Minion(cmdbw cmdb.Client, pusher push.Pusher) MinionService {
	return &minionService{
		cmdbw:  cmdbw,
		pusher: pusher,
	}
}

type minionService struct {
	cmdbw  cmdb.Client
	pusher push.Pusher
}

func (biz *minionService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope, likes []gen.Condition) (int64, []*param.MinionSummary) {
	tagTbl := query.MinionTag
	monTbl := query.Minion
	dao := monTbl.WithContext(ctx).
		Distinct(monTbl.ID).
		LeftJoin(tagTbl, monTbl.ID.EqCol(tagTbl.MinionID)).
		Order(monTbl.ID)
	if len(likes) != 0 {
		for i, like := range likes {
			likes[i] = dao.Or(like)
		}
		dao.Where(likes...)
	}
	db := dao.UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}
	var monIDs []int64
	if db.Scopes(page.DBScope(count)).
		Scan(&monIDs); len(monIDs) == 0 {
		return 0, nil
	}
	// 查询数据
	minions, err := monTbl.WithContext(ctx).
		Where(monTbl.ID.In(monIDs...)).
		Find()
	if err != nil {
		return 0, nil
	}

	tagMap := map[int64][]string{}
	infoMap := map[int64]*model.SysInfo{}

	if tags, _ := tagTbl.WithContext(ctx).
		Where(tagTbl.MinionID.In(monIDs...)).
		Find(); len(tags) != 0 {
		tagMap = model.MinionTags(tags).ToMap()
	}
	infoTbl := query.SysInfo
	if infos, _ := infoTbl.WithContext(ctx).Where(infoTbl.ID.In(monIDs...)).Find(); len(infos) != 0 {
		infoMap = model.SysInfos(infos).ToMap()
	}

	ret := make([]*param.MinionSummary, 0, len(monIDs))
	for _, m := range minions {
		id := m.ID
		ms := &param.MinionSummary{
			ID:         id,
			Inet:       m.Inet,
			Goos:       m.Goos,
			Edition:    m.Edition,
			Status:     m.Status,
			IDC:        m.IDC,
			IBu:        m.IBu,
			Comment:    m.Comment,
			BrokerName: m.BrokerName,
			Unload:     m.Unload,
			Uptime:     m.Uptime.Time,
			Customized: m.Customized,
			Unstable:   m.Unstable,
			Tags:       tagMap[id],
		}
		if ms.Tags == nil {
			ms.Tags = []string{}
		}
		if inf := infoMap[id]; inf != nil {
			ms.CPUCore = inf.CPUCore
			ms.MemTotal = inf.MemTotal
			ms.MemFree = inf.MemFree
		}
		ret = append(ret, ms)
	}

	return count, ret
}

func (biz *minionService) Detail(ctx context.Context, id int64) (*param.MinionDetail, error) {
	tbl := query.Minion
	dat := new(param.MinionDetail)
	if err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Scan(dat); err != nil {
		return nil, err
	}

	tagTbl := query.MinionTag
	dat.Tags, _ = tagTbl.WithContext(ctx).Where(tagTbl.MinionID.Eq(id)).Find()
	if dat.Tags == nil {
		dat.Tags = []*model.MinionTag{}
	}

	infoTbl := query.SysInfo
	info, _ := infoTbl.WithContext(ctx).Where(infoTbl.ID.Eq(id)).First()
	if info != nil {
		dat.Release = info.Release
		dat.CPUCore = info.CPUCore
		dat.MemTotal = info.MemTotal
		dat.MemFree = info.MemFree
		dat.SwapTotal = info.SwapTotal
		dat.SwapFree = info.SwapFree
		dat.HostID = info.HostID
		dat.Family = info.Family
		dat.BootAt = info.BootAt
		dat.VirtualSys = info.Virtual
		dat.VirtualRole = info.VirtualRole
		dat.ProcNumber = info.ProcNumber
		dat.Hostname = info.Hostname
		dat.KernelVersion = info.KernelVersion
		dat.AgentTotal = info.AgentTotal
		dat.AgentAlloc = info.AgentAlloc
	}

	return dat, nil
}

func (biz *minionService) Drop(ctx context.Context, id int64) error {
	tbl := query.Minion
	mon, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if mon.Status != model.MSDelete {
		return errcode.ErrDeleteFailed
	}

	// 查询该节点关联的标签
	var tags []string
	tagTbl := query.MinionTag
	if err = tagTbl.WithContext(ctx).
		Distinct(tagTbl.Tag).
		Where(tagTbl.MinionID.Eq(id)).
		Scan(&tags); err != nil {
		return err
	}

	subTbl := query.Substance
	if err = query.Q.Transaction(func(tx *query.Query) error {
		if _, exx := tx.WithContext(ctx).MinionTag.
			Where(tagTbl.MinionID.Eq(id)).Delete(); exx != nil {
			return exx
		}
		if _, exx := tx.WithContext(ctx).Substance.
			Where(subTbl.MinionID.Eq(id)).Delete(); exx != nil {
			return exx
		}

		_, exx := tx.WithContext(ctx).Minion.Where(tbl.ID.Eq(id)).Delete()
		return exx
	}); err != nil {
		return err
	}

	cmdbTbl := query.Cmdb
	_, _ = cmdbTbl.WithContext(ctx).Where(cmdbTbl.ID.Eq(id)).Delete()
	infTbl := query.SysInfo
	_, _ = infTbl.WithContext(ctx).Where(infTbl.ID.Eq(id)).Delete()
	evtTbl := query.Event
	_, _ = evtTbl.WithContext(ctx).Where(evtTbl.MinionID.Eq(id)).Delete()
	rskTbl := query.Risk
	_, _ = rskTbl.WithContext(ctx).Where(rskTbl.MinionID.Eq(id)).Delete()

	accTbl := query.MinionAccount
	_, _ = accTbl.WithContext(ctx).Where(accTbl.MinionID.Eq(id)).Delete()
	grpTbl := query.MinionGroup
	_, _ = grpTbl.WithContext(ctx).Where(grpTbl.MinionID.Eq(id)).Delete()
	lisTbl := query.MinionListen
	_, _ = lisTbl.WithContext(ctx).Where(lisTbl.MinionID.Eq(id)).Delete()
	lonTbl := query.MinionListen
	_, _ = lonTbl.WithContext(ctx).Where(lonTbl.MinionID.Eq(id)).Delete()
	procTbl := query.MinionProcess
	_, _ = procTbl.WithContext(ctx).Where(procTbl.MinionID.Eq(id)).Delete()
	taskTbl := query.MinionTask
	_, _ = taskTbl.WithContext(ctx).Where(taskTbl.MinionID.Eq(id)).Delete()

	// 清理该节点的 SBOM 信息
	bomMonTbl := query.SBOMMinion
	_, _ = bomMonTbl.WithContext(ctx).Where(bomMonTbl.ID.Eq(id)).Delete()
	bomPjtTbl := query.SBOMProject
	_, _ = bomPjtTbl.WithContext(ctx).Where(bomPjtTbl.MinionID.Eq(id)).Delete()
	bomComTbl := query.SBOMComponent
	_, _ = bomComTbl.WithContext(ctx).Where(bomComTbl.MinionID.Eq(id)).Delete()

	size := len(tags)
	if size == 0 {
		return nil
	}
	// -----[ 删除野标签 ]-----
	var afterTags []string
	if err = tagTbl.WithContext(ctx).
		Distinct(tagTbl.Tag).
		Where(tagTbl.Tag.In(tags...)).
		Scan(&afterTags); err != nil {
		return err
	}
	thm := make(map[string]struct{}, size)
	for _, tag := range tags {
		thm[tag] = struct{}{}
	}
	for _, tag := range afterTags {
		delete(thm, tag)
	}
	wildTags := make([]string, 0, 10)
	for tag := range thm {
		wildTags = append(wildTags, tag)
	}
	if len(wildTags) == 0 {
		return nil
	}
	// 删除 effect
	effTbl := query.Effect
	_, err = effTbl.WithContext(ctx).Where(effTbl.Tag.In(wildTags...)).Delete()

	return err
}

func (biz *minionService) Create(ctx context.Context, mc *param.MinionCreate) error {
	inet := net.ParseIP(mc.Inet)
	if len(inet) == 0 || inet.IsLoopback() || inet.IsUnspecified() || inet.Equal(net.IPv4bcast) {
		return errcode.ErrInetAddress
	}

	// 检查IPv4是否重复
	tbl := query.Minion
	ipv4 := inet.String()
	if count, err := tbl.WithContext(ctx).
		Where(tbl.Inet.Eq(ipv4)).
		Count(); err != nil || count != 0 {
		return errcode.FmtErrInetExist.Fmt(ipv4)
	}

	mon := &model.Minion{
		Inet:   ipv4,
		Goos:   mc.Goos,
		Arch:   mc.Arch,
		Status: model.MSOffline,
	}
	if err := tbl.WithContext(ctx).Create(mon); err != nil {
		return err
	}
	tags := []*model.MinionTag{{Tag: ipv4, MinionID: mon.ID, Kind: model.TkLifelong}}
	if mc.Goos != "" {
		tags = append(tags, &model.MinionTag{Tag: mc.Goos, MinionID: mon.ID, Kind: model.TkLifelong})
	}
	if mc.Arch != "" {
		tags = append(tags, &model.MinionTag{Tag: mc.Arch, MinionID: mon.ID, Kind: model.TkLifelong})
	}

	_ = query.MinionTag.WithContext(ctx).Create(tags...)
	_ = biz.cmdbw.FetchAndSave(ctx, mon.ID, ipv4)

	return nil
}

func (biz *minionService) Delete(ctx context.Context, scope dynsql.Scope, likes []gen.Condition) error {
	cbFunc := func(ctx context.Context, bid int64, mids []int64) error {
		tbl := query.Minion
		deleted := uint8(model.MSDelete)
		_, _ = tbl.WithContext(ctx).
			Where(tbl.Status.Neq(deleted), tbl.ID.In(mids...)).
			UpdateColumnSimple(tbl.Status.Value(deleted))
		// 通知 broker 节点下线
		biz.pusher.Command(ctx, bid, mids, "offline")
		return nil
	}

	err := biz.batchFunc(scope, likes, cbFunc)

	return err
}

func (biz *minionService) CSV(ctx context.Context) sheet.CSVStreamer {
	read := sheet.MinionCSV(ctx, 500, true)
	return sheet.NewCSV(read)
}

func (biz *minionService) Upgrade(ctx context.Context, mid, binID int64) error {
	tbl := query.Minion
	mon, err := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Status, tbl.BrokerID).
		Where(tbl.ID.Eq(mid)).First()
	if err != nil {
		return err
	}
	if mon.Status == model.MSDelete {
		return errcode.ErrNodeStatus
	}

	binTbl := query.MinionBin
	bin, err := binTbl.WithContext(ctx).Where(binTbl.ID.Eq(binID)).First()
	if err != nil {
		return err
	}
	if bin.Deprecated {
		return errcode.ErrDeprecated
	}
	if bin.Unstable {
		return errcode.ErrReleaseUnstable
	}

	semver := string(bin.Semver)
	biz.pusher.Upgrade(ctx, mon.BrokerID, []int64{mid}, semver, bin.Customized)

	return nil
}

func (biz *minionService) Batch(ctx context.Context, scope dynsql.Scope, likes []gen.Condition, cmd string) error {
	cbFunc := func(ctx context.Context, bid int64, mids []int64) error {
		// resync restart upgrade offline
		biz.pusher.Command(ctx, bid, mids, cmd)
		return nil
	}

	go biz.batchFunc(scope, likes, cbFunc)

	return nil
}

func (biz *minionService) Command(ctx context.Context, mid int64, cmd string) error {
	tbl := query.Minion
	mon, err := tbl.WithContext(ctx).
		Select(tbl.Status, tbl.BrokerID).
		Where(tbl.ID.Eq(mid)).
		First()
	if err != nil {
		return err
	}
	status := mon.Status
	if status == model.MSDelete || status == model.MSInactive {
		return errcode.ErrNodeStatus
	}

	biz.pusher.Command(ctx, mon.BrokerID, []int64{mid}, cmd)

	return nil
}

func (biz *minionService) Unload(ctx context.Context, mid int64, unload bool) error {
	// 查询节点信息
	tbl := query.Minion
	mon, err := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Status, tbl.BrokerID, tbl.Unload, tbl.Inet).
		Where(tbl.ID.Eq(mid)).
		First()
	if err != nil {
		return err
	}
	status := mon.Status
	if status == model.MSDelete || status == model.MSInactive {
		return errcode.ErrNodeStatus
	}
	if unload == mon.Unload {
		return nil
	}

	_, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(mid)).
		UpdateColumnSimple(tbl.Unload.Value(unload))
	if err == nil {
		biz.pusher.TaskSync(ctx, mon.BrokerID, []int64{mid})
	}

	return err
}

func (biz *minionService) BatchTag(ctx context.Context, scope dynsql.Scope, likes []gen.Condition, creates, deletes []string) error {
	fn := func(ctx context.Context, bid int64, mids []int64) error {
		err := query.Q.Transaction(func(tx *query.Query) error {
			ll := int8(model.TkLifelong)
			tbl := tx.MinionTag
			dao := tbl.WithContext(ctx)
			for _, mid := range mids {
				if len(deletes) != 0 {
					_, _ = dao.Where(tbl.MinionID.Eq(mid), tbl.Kind.Neq(ll), tbl.Tag.In(deletes...)).Delete()
				}
				if size := len(creates); size != 0 {
					dats := make([]*model.MinionTag, 0, size)
					for _, tag := range creates {
						dats = append(dats, &model.MinionTag{Tag: tag, MinionID: mid, Kind: model.TkManual})
					}
					_ = dao.Clauses(clause.OnConflict{DoNothing: true}).Create(dats...)
				}
			}

			biz.pusher.TaskSync(ctx, bid, mids)

			return nil
		})

		return err
	}

	return biz.batchFunc(scope, likes, fn)
}

func (biz *minionService) batchFunc(
	scope dynsql.Scope,
	likes []gen.Condition,
	cb func(ctx context.Context, bid int64, mids []int64) error,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	tbl, tagTbl := query.Minion, query.MinionTag
	deleted := uint8(model.MSDelete)
	dao := tbl.WithContext(ctx).
		Distinct(tbl.ID).
		LeftJoin(tagTbl, tagTbl.MinionID.EqCol(tbl.ID)).
		Order(tbl.ID)
	if len(likes) != 0 {
		for i, like := range likes {
			likes[i] = dao.Or(like)
		}
		dao.Where(likes...)
	}

	limit, offset := 200, 0
	db := dao.UnderlyingDB().
		Where(tbl.Status.Neq(deleted)).
		Scopes(scope.Where).
		Limit(limit)

	var err error
	for err == nil {
		var mids []int64
		err = db.Offset(offset).Limit(limit).Scan(&mids).Error
		size := len(mids)
		if err != nil || size == 0 {
			break
		}
		offset += size

		dats, exx := tbl.WithContext(ctx).
			Select(tbl.ID, tbl.Status, tbl.BrokerID).
			Where(tbl.ID.In(mids...)).
			Find()
		if exx != nil || len(dats) == 0 {
			err = exx
			break
		}

		hm := make(map[int64][]int64, 16)
		for _, dat := range dats {
			bid := dat.BrokerID
			if dat.Status == model.MSDelete || bid == 0 {
				continue
			}
			hm[bid] = append(hm[bid], dat.ID)
		}

		for bid, ids := range hm {
			if err = cb(ctx, bid, ids); err != nil {
				break
			}
		}
		if size < limit {
			break
		}
	}

	return err
}
