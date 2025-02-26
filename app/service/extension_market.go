package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
	"gorm.io/gen/field"
)

func NewExtensionMarket(qry *query.Query) *ExtensionMarket {
	return &ExtensionMarket{
		qry: qry,
	}
}

type ExtensionMarket struct {
	qry *query.Query
}

func (mkt *ExtensionMarket) Page(ctx context.Context, page mrequest.Pager, category string) (int64, []*model.ExtensionMarket) {
	tbl := mkt.qry.ExtensionMarket
	dao := tbl.WithContext(ctx)
	if category != "" {
		dao = dao.Where(tbl.Category.Eq(category))
	}

	if kw := page.Keyword(); kw != "" {
		like := field.Or(tbl.Name.Like(kw), tbl.Intro.Like(kw))
		dao = dao.Where(like)
	}

	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (mkt *ExtensionMarket) Create(ctx context.Context, req *mrequest.ExtensionMarketCreate, cu *session.Ident) (*model.ExtensionMarket, error) {
	now := time.Now()
	content := req.Content
	sum := sha1.Sum([]byte(content))
	contentSHA1 := hex.EncodeToString(sum[:])
	const firstVersion = 1

	createdBy := model.Operator{ID: cu.ID, Username: cu.Username, Nickname: cu.Nickname}
	market := &model.ExtensionMarket{
		Name:        req.Name,
		Intro:       req.Intro,
		Category:    req.Category,
		Version:     firstVersion,
		Content:     content,
		ContentSHA1: contentSHA1,
		Changelog:   req.Changelog,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	record := &model.ExtensionRecord{
		Version:     firstVersion,
		CreatedBy:   createdBy,
		Content:     content,
		ContentSHA1: contentSHA1,
		CreatedAt:   now,
	}

	if err := mkt.qry.Transaction(func(tx *query.Query) error {
		if err := tx.ExtensionMarket.
			WithContext(ctx).
			Create(market); err != nil {
			return err
		}
		record.ExtensionID = market.ID

		return tx.ExtensionRecord.
			WithContext(ctx).
			Create(record)
	}); err != nil {
		return nil, err
	}

	return market, nil
}

func (mkt *ExtensionMarket) Update(ctx context.Context, req *mrequest.ExtensionMarketUpdate, cu *session.Ident) error {
	now := time.Now()
	tbl := mkt.qry.ExtensionMarket
	data, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		First()
	if err != nil {
		return err
	}

	content := req.Content
	sum := sha1.Sum([]byte(content))
	contentSHA1 := hex.EncodeToString(sum[:])
	changed := data.ContentSHA1 != contentSHA1

	updatedBy := model.Operator{ID: cu.ID, Username: cu.Username, Nickname: cu.Nickname}
	columns := []field.AssignExpr{
		tbl.Intro.Value(req.Intro),
		tbl.Changelog.Value(req.Changelog),
		tbl.UpdatedBy.Value(updatedBy),
		tbl.UpdatedAt.Value(now),
	}
	modifiedVersion := data.Version
	if changed {
		modifiedVersion = data.Version + 1
		columns = append(columns,
			tbl.Content.Value(content),
			tbl.ContentSHA1.Value(contentSHA1),
			tbl.Version.Value(modifiedVersion),
		)
	}

	return mkt.qry.Transaction(func(tx *query.Query) error {
		mktTbl := tx.ExtensionMarket
		if _, exx := mktTbl.WithContext(ctx).
			Where(tbl.ID.Eq(req.ID), tbl.Version.Eq(data.Version)).
			UpdateSimple(columns...); exx != nil || !changed {
			return exx
		}

		record := &model.ExtensionRecord{
			ExtensionID: req.ID,
			Version:     modifiedVersion,
			CreatedBy:   updatedBy,
			Content:     content,
			ContentSHA1: contentSHA1,
			CreatedAt:   now,
		}
		rcdTbl := tx.ExtensionRecord

		return rcdTbl.WithContext(ctx).Create(record)
	})
}

func (mkt *ExtensionMarket) Delete(ctx context.Context, id int64) error {
	return mkt.qry.Transaction(func(tx *query.Query) error {
		mktTbl := mkt.qry.ExtensionMarket
		if _, err := mktTbl.WithContext(ctx).
			Where(mktTbl.ID.Eq(id)).
			Delete(); err != nil {
			return err
		}

		rcdTbl := mkt.qry.ExtensionRecord
		_, err := rcdTbl.WithContext(ctx).
			Where(rcdTbl.ExtensionID.Eq(id)).
			Delete()

		return err
	})
}

func (mkt *ExtensionMarket) Records(ctx context.Context, id int64) ([]*model.ExtensionRecord, error) {
	tbl := mkt.qry.ExtensionRecord
	return tbl.WithContext(ctx).
		Where(tbl.ExtensionID.Eq(id)).
		Order(tbl.Version.Desc()).
		Limit(1000).Find()
}
