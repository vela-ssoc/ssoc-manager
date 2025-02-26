package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
)

type DashService interface {
	Status(ctx context.Context) *mrequest.DashStatusResp
	Goos(ctx context.Context) *mrequest.DashGoosVO
	Edition(ctx context.Context) []*mrequest.DashEditionVO
	Evtlvl(ctx context.Context) *mrequest.DashELevelResp
	Risklvl(ctx context.Context) *mrequest.DashRLevelResp
	Risksts(ctx context.Context) *mrequest.DashRiskstsResp
}

func Dash(qry *query.Query) DashService {
	return &dashService{
		qry: qry,
	}
}

type dashService struct {
	qry *query.Query
}

func (biz *dashService) Status(ctx context.Context) *mrequest.DashStatusResp {
	var tmp []*struct {
		Status model.MinionStatus `gorm:"column:status"`
		Count  int                `gorm:"column:count"`
	}

	tbl := biz.qry.Minion
	tbl.WithContext(ctx).UnderlyingDB().Select("status", "COUNT(*) AS count").
		Model(&model.Minion{}).Group("status").Scan(&tmp)

	ret := new(mrequest.DashStatusResp)
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

func (biz *dashService) Goos(ctx context.Context) *mrequest.DashGoosVO {
	ql := "SELECT COUNT(IF(goos = 'linux', TRUE, NULL))   AS linux,   " +
		"         COUNT(IF(goos = 'windows', TRUE, NULL)) AS windows, " +
		"         COUNT(IF(goos = 'darwin', TRUE, NULL))  AS darwin   " +
		"FROM minion;"
	ret := new(mrequest.DashGoosVO)
	biz.qry.Minion.WithContext(ctx).UnderlyingDB().Raw(ql).Scan(&ret)

	return ret
}

func (biz *dashService) Edition(ctx context.Context) []*mrequest.DashEditionVO {
	var dats []*mrequest.DashEditionVO
	biz.qry.Minion.WithContext(ctx).UnderlyingDB().
		Select("edition", "COUNT(*) AS total").
		Group("edition").
		Order("edition DESC"). // 按照版本号降序
		Scan(&dats)

	return dats
}

func (biz *dashService) Evtlvl(ctx context.Context) *mrequest.DashELevelResp {
	var tmp []*struct {
		Level model.EventLevel `gorm:"column:level"`
		Count int              `gorm:"column:count"`
	}

	biz.qry.Event.WithContext(ctx).UnderlyingDB().
		Select("level", "COUNT(*) AS count").
		Group("level").
		Scan(&tmp)

	var res mrequest.DashELevelResp
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

func (biz *dashService) Risklvl(ctx context.Context) *mrequest.DashRLevelResp {
	var tmp []*struct {
		Level model.RiskLevel `gorm:"column:level"`
		Count int             `gorm:"column:count"`
	}
	biz.qry.Risk.WithContext(ctx).UnderlyingDB().
		Select("level", "COUNT(*) AS count").
		Group("level").
		Scan(&tmp)

	var res mrequest.DashRLevelResp
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

func (biz *dashService) Risksts(ctx context.Context) *mrequest.DashRiskstsResp {
	var tmp []*struct {
		Status model.RiskStatus `gorm:"column:status"`
		Count  int              `gorm:"column:count"`
	}

	biz.qry.Risk.WithContext(ctx).UnderlyingDB().
		Select("status", "COUNT(*) AS count").
		Group("status").
		Scan(&tmp)

	var res mrequest.DashRiskstsResp
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
