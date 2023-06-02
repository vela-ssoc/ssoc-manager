package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"gorm.io/gorm/clause"
)

type RiskIPService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.RiskIP)
	Delete(ctx context.Context, ids []int64) error
	Create(ctx context.Context, rc *param.RiskIPCreate) error
	Update(ctx context.Context, rc *param.RiskIPUpdate) error
	Import(ctx context.Context, rii *param.RiskIPImport) error
}

func RiskIP() RiskIPService {
	return &riskIPService{}
}

type riskIPService struct{}

func (biz *riskIPService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.RiskIP) {
	tbl := query.RiskIP
	db := tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.RiskIP
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (biz *riskIPService) Delete(ctx context.Context, ids []int64) error {
	tbl := query.RiskIP
	_, err := tbl.WithContext(ctx).Where(tbl.ID.In(ids...)).Delete()
	return err
}

func (biz *riskIPService) Create(ctx context.Context, rc *param.RiskIPCreate) error {
	dats := rc.Models()
	tbl := query.RiskIP
	return tbl.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(dats...)
}

func (biz *riskIPService) Update(ctx context.Context, rc *param.RiskIPUpdate) error {
	dat := &model.RiskIP{
		ID:       rc.ID,
		IP:       rc.IP,
		Kind:     rc.Kind,
		Origin:   rc.Origin,
		BeforeAt: rc.BeforeAt,
	}
	tbl := query.RiskIP
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(rc.ID)).
		Updates(dat)

	return err
}

func (biz *riskIPService) Import(ctx context.Context, rii *param.RiskIPImport) error {
	mods := rii.Models()
	conflict := clause.OnConflict{DoNothing: true} // 默认冲突时跳过
	if rii.Update {                                // 如果规则冲突时执行更新操作
		conflict.DoNothing = false
		conflict.DoUpdates = clause.AssignmentColumns([]string{"origin", "before_at", "updated_at"})
	}
	tbl := query.RiskIP

	return tbl.WithContext(ctx).
		Clauses(conflict).
		CreateInBatches(mods, 100)
}
