package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type DashService interface {
	Status(ctx context.Context) *param.DashStatusResp
	Goos(ctx context.Context) *param.DashGoosVO
	Edition(ctx context.Context) []*param.DashEditionVO
	Evtlvl(ctx context.Context) *param.DashELevelResp
	Risklvl(ctx context.Context) *param.DashRLevelResp
	Risksts(ctx context.Context) *param.DashRiskstsResp
}

func Dash(qry *query.Query) DashService {
	return &dashService{
		qry: qry,
	}
}

type dashService struct {
	qry *query.Query
}

func (biz *dashService) Status(ctx context.Context) *param.DashStatusResp {
	var tmp []*struct {
		Status model.MinionStatus `gorm:"column:status"`
		Count  int                `gorm:"column:count"`
	}

	tbl := biz.qry.Minion
	tbl.WithContext(ctx).UnderlyingDB().Select("status", "COUNT(*) AS count").
		Model(&model.Minion{}).Group("status").Scan(&tmp)

	ret := new(param.DashStatusResp)
	for _, tp := range tmp {
		switch tp.Status {
		case model.MSInactive:
			ret.Inactive = tp.Count
		case model.MSOffline:
			ret.Offline = tp.Count
		case model.MSOnline:
			ret.Online = tp.Count
		case model.MSDelete:
			ret.Deleted = tp.Count
		}
	}

	return ret
}

func (biz *dashService) Goos(ctx context.Context) *param.DashGoosVO {
	ql := "SELECT COUNT(IF(goos = 'linux', TRUE, NULL))   AS linux,   " +
		"         COUNT(IF(goos = 'windows', TRUE, NULL)) AS windows, " +
		"         COUNT(IF(goos = 'darwin', TRUE, NULL))  AS darwin   " +
		"FROM minion;"
	ret := new(param.DashGoosVO)
	biz.qry.Minion.WithContext(ctx).UnderlyingDB().Raw(ql).Scan(&ret)

	return ret
}

func (biz *dashService) Edition(ctx context.Context) []*param.DashEditionVO {
	var dats []*param.DashEditionVO
	biz.qry.Minion.WithContext(ctx).UnderlyingDB().
		Select("edition", "COUNT(*) AS total").
		Group("edition").
		Order("INET_ATON(CONCAT(edition, '.0')) DESC"). // 按照版本号降序
		Scan(&dats)

	return dats
}

func (biz *dashService) Evtlvl(ctx context.Context) *param.DashELevelResp {
	var tmp []*struct {
		Level model.EventLevel `gorm:"column:level"`
		Count int              `gorm:"column:count"`
	}

	biz.qry.Event.WithContext(ctx).UnderlyingDB().
		Select("level", "COUNT(*) AS count").
		Group("level").
		Scan(&tmp)

	var res param.DashELevelResp
	for _, tp := range tmp {
		switch tp.Level {
		case model.ELvlCritical:
			res.Critical = tp.Count
		case model.ELvlMajor:
			res.Major = tp.Count
		case model.ELvlMinor:
			res.Minor = tp.Count
		case model.ELvlNote:
			res.Note = tp.Count
		}
	}

	return &res
}

func (biz *dashService) Risklvl(ctx context.Context) *param.DashRLevelResp {
	var tmp []*struct {
		Level model.RiskLevel `gorm:"column:level"`
		Count int             `gorm:"column:count"`
	}
	biz.qry.Risk.WithContext(ctx).UnderlyingDB().
		Select("level", "COUNT(*) AS count").
		Group("level").
		Scan(&tmp)

	var res param.DashRLevelResp
	for _, tp := range tmp {
		switch tp.Level {
		case model.RLvlCritical:
			res.Critical = tp.Count
		case model.RLvlHigh:
			res.High = tp.Count
		case model.RLvlMiddle:
			res.Middle = tp.Count
		case model.RLvlLow:
			res.Low = tp.Count
		}
	}

	return &res
}

func (biz *dashService) Risksts(ctx context.Context) *param.DashRiskstsResp {
	var tmp []*struct {
		Status model.RiskStatus `gorm:"column:status"`
		Count  int              `gorm:"column:count"`
	}

	biz.qry.Risk.WithContext(ctx).UnderlyingDB().
		Select("status", "COUNT(*) AS count").
		Group("status").
		Scan(&tmp)

	var res param.DashRiskstsResp
	for _, tp := range tmp {
		switch tp.Status {
		case model.RSUnprocessed:
			res.Unprocessed = tp.Count
		case model.RSProcessed:
			res.Processed = tp.Count
		case model.RSIgnore:
			res.Ignore = tp.Count
		}
	}

	return &res
}

func (biz *dashService) BGoos(ctx context.Context, page, size int) {
}
