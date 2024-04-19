package transact

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
)

func EffectTaskTx(_ context.Context, taskID int64, tags []string) (brokerIDs []int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	et := &effectTaskTx{
		ctx:    ctx,
		taskID: taskID,
		tags:   tags,
		limit:  100,
		bids:   make([]int64, 0, 16),
	}

	if err = query.Q.Transaction(et.Func); err != nil {
		return nil, err
	}

	return et.bids, nil
}

type effectTaskTx struct {
	ctx    context.Context
	taskID int64
	tags   []string
	limit  int
	bids   []int64
}

func (et *effectTaskTx) Func(tx *query.Query) error {
	now := time.Now()
	ctx := et.ctx
	limit, offset := et.limit, 0
	tagTbl := query.MinionTag
	monTbl := query.Minion
	bmap := make(map[int64]struct{}, 16)

	for {
		minionIDs := make([]int64, 0, limit)
		err := tx.MinionTag.WithContext(ctx).
			Distinct(tagTbl.MinionID).
			Where(tagTbl.Tag.In(et.tags...)).
			Order(tagTbl.MinionID).
			Limit(et.limit).
			Offset(offset).
			Scan(&minionIDs)
		if err != nil {
			return err
		}

		size := len(minionIDs)
		if size == 0 {
			break
		}
		offset += size

		// 查询 broker_id 与 broker_name
		minions, err := tx.Minion.WithContext(ctx).
			Select(monTbl.ID, monTbl.Inet, monTbl.BrokerID, monTbl.BrokerName).
			Where(monTbl.BrokerID.Neq(0)).
			Where(monTbl.ID.In(minionIDs...)).
			Find()
		if err != nil {
			return err
		}

		num := len(minions)
		if num == 0 {
			break
		}

		tasks := make([]*model.SubstanceTask, 0, limit)
		for _, mon := range minions {
			bid := mon.BrokerID
			if _, ok := bmap[bid]; !ok {
				bmap[bid] = struct{}{}
				et.bids = append(et.bids, bid)
			}
			task := &model.SubstanceTask{
				TaskID:     et.taskID,
				MinionID:   mon.ID,
				Inet:       mon.Inet,
				BrokerID:   bid,
				BrokerName: mon.BrokerName,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			tasks = append(tasks, task)
		}

		if err = tx.WithContext(ctx).SubstanceTask.
			CreateInBatches(tasks, limit); err != nil {
			return err
		}
	}

	return nil
}
