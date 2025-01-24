package param

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type EffectCreate struct {
	Name       string   `json:"name"       validate:"required,lte=50"`                    // 配置发布名字
	Enable     bool     `json:"enable"`                                                   // 是否开启
	Tags       []string `json:"tags"       validate:"gte=1,lte=100,unique,dive,required"` // 生效的节点 tag
	Exclusion  []string `json:"exclusion"  validate:"lte=100,unique,dive,ipv4"`           // 排除的节点 (以节点 IPv4 维度)
	Substances Int64s   `json:"substances" validate:"gte=1,lte=100,unique"`               // 配置
}

func (ec EffectCreate) Check(ctx context.Context, qry *query.Query) error {
	// 1. 标签必须已经存在
	tsz := len(ec.Tags)
	tagTbl := qry.MinionTag
	count, _ := tagTbl.WithContext(ctx).
		Distinct(tagTbl.Tag).
		Where(tagTbl.Tag.In(ec.Tags...)).
		Count()
	if int(count) != tsz {
		return errcode.ErrTagNotExist
	}

	// 2. 配置必须已经存在且全部为公有配置
	if size := len(ec.Substances); size != 0 {
		subTbl := qry.Substance
		count, _ = subTbl.WithContext(ctx).
			Where(subTbl.MinionID.Eq(0)).
			Where(subTbl.ID.In(ec.Substances...)).
			Count()
		if int(count) != size {
			return errcode.ErrSubstanceNotExist
		}
	}

	return nil
}

func (ec EffectCreate) Expand(subID, createdID int64) []*model.Effect {
	ret := make([]*model.Effect, 0, 32)
	now := time.Now()

	for _, tag := range ec.Tags {
		for _, sub := range ec.Substances {
			eff := &model.Effect{
				Name:      ec.Name,
				SubmitID:  subID,
				Tag:       tag,
				EffectID:  sub,
				Enable:    ec.Enable,
				Exclusion: ec.Exclusion,
				CreatedID: createdID,
				UpdatedID: createdID,
				CreatedAt: now,
				UpdatedAt: now,
			}
			ret = append(ret, eff)
		}
	}

	return ret
}

type EffectUpdate struct {
	EffectCreate
	IntID
	Version int64 `json:"version"`
}

func (ec EffectUpdate) Expand(reduce *model.EffectReduce, updatedID int64) []*model.Effect {
	ret := make([]*model.Effect, 0, 32)
	now := time.Now()
	subID := reduce.SubmitID
	version := ec.Version + 1

	for _, tag := range ec.Tags {
		for _, sub := range ec.Substances {
			eff := &model.Effect{
				Name:      ec.Name,
				SubmitID:  subID,
				Tag:       tag,
				EffectID:  sub,
				Enable:    ec.Enable,
				Exclusion: ec.Exclusion,
				CreatedID: reduce.CreatedID,
				UpdatedID: updatedID,
				CreatedAt: reduce.CreatedAt,
				UpdatedAt: now,
				Version:   version,
			}
			ret = append(ret, eff)
		}
	}

	return ret
}

type EffectTaskResp struct {
	TaskID int64 `json:"task_id,string"`
}

type EffectSummary struct {
	ID         int64     `json:"id,string"`
	Name       string    `json:"name"`
	Tags       []string  `json:"tags"`
	Enable     bool      `json:"enable"`
	Version    int64     `json:"version"`
	Exclusion  []string  `json:"exclusion"`
	Compounds  []*IDName `json:"compounds"`
	Substances []*IDName `json:"substances"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type EffectProgress struct {
	ID       int64 `json:"id,string"` // 任务 ID
	Count    int   `json:"count"`     // 总数
	Executed int   `json:"executed"`  // 已经下发完毕的
	Failed   int   `json:"failed"`    // 下发失败的
}

type EffectProgressesRequest struct {
	OptionalID
	Page
}
