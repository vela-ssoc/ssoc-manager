package service

import (
	"context"
	"net"

	"gorm.io/gen/field"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-common-mb/integration/cmdb"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/internal/sheet"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type MinionService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.MinionSummary)
	Detail(ctx context.Context, id int64) (*param.MinionDetail, error)
	Drop(ctx context.Context, id int64) error
	Create(ctx context.Context, mc *param.MinionCreate) error
	Delete(ctx context.Context, scope dynsql.Scope) error
	Activate(ctx context.Context, id []int64) error
	CSV(ctx context.Context) sheet.CSVStreamer
}

func Minion(cmdbw cmdb.Client) MinionService {
	return &minionService{
		cmdbw: cmdbw,
	}
}

type minionService struct {
	cmdbw cmdb.Client
}

func (biz *minionService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.MinionSummary) {
	tagTbl := query.MinionTag
	monTbl := query.Minion

	db := monTbl.WithContext(ctx).
		Distinct(monTbl.ID).
		LeftJoin(tagTbl, monTbl.ID.EqCol(tagTbl.MinionID)).
		Order(monTbl.ID).
		UnderlyingDB().
		Scopes(scope.Where)
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
	minions, err := monTbl.WithContext(ctx).Where(monTbl.ID.In(monIDs...)).Find()
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
			ID:      id,
			Inet:    m.Inet,
			Goos:    m.Goos,
			Edition: m.Edition,
			Status:  m.Status,
			IDC:     m.IDC,
			IBu:     m.IBu,
			Comment: m.Comment,
			Tags:    tagMap[id],
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
	monTbl := query.Minion
	infoTbl := query.SysInfo
	dat := new(param.MinionDetail)
	err := monTbl.WithContext(ctx).
		Select(field.ALL).
		LeftJoin(infoTbl, infoTbl.ID.EqCol(monTbl.ID)).
		Where(monTbl.ID.Eq(id)).
		Scan(&dat)
	if err != nil {
		return nil, err
	}

	tagTbl := query.MinionTag
	dat.Tags, _ = tagTbl.WithContext(ctx).Where(tagTbl.MinionID.Eq(id)).Find()
	if dat.Tags == nil {
		dat.Tags = []*model.MinionTag{}
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

func (biz *minionService) Delete(ctx context.Context, scope dynsql.Scope) error {
	tbl := query.Minion
	tagTbl := query.MinionTag

	deleted := uint8(model.MSDelete)
	db := tbl.WithContext(ctx).
		Distinct(tbl.ID).
		Where(tbl.Status.Neq(deleted)).
		LeftJoin(tagTbl, tbl.ID.EqCol(tagTbl.MinionID)).
		Order(tbl.ID).
		UnderlyingDB().
		Scopes(scope.Where).
		Limit(100)

	var round int
	var err error
	for round < 10 {
		round++
		var mids []int64
		if err = db.Scan(&mids).Error; err != nil || len(mids) == 0 {
			break
		}
		_, _ = tbl.WithContext(ctx).
			Where(tbl.Status.Neq(deleted), tbl.ID.In(mids...)).
			UpdateColumnSimple(tbl.Status.Value(deleted))
		// 通知 broker 节点下线
	}

	return err
}

func (biz *minionService) Activate(ctx context.Context, id []int64) error {
	inactive, offline := uint8(model.MSInactive), uint8(model.MSOffline)
	tbl := query.Minion
	_, err := tbl.WithContext(ctx).
		Where(tbl.Status.Eq(inactive)).
		UpdateColumnSimple(tbl.Status.Value(offline))
	return err
}

func (biz *minionService) CSV(ctx context.Context) sheet.CSVStreamer {
	read := sheet.MinionCSV(ctx, 500, true)
	return sheet.NewCSV(read)
}
