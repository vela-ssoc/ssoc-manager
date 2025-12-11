package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/vela-ssoc/luatemplate"
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
)

func NewSubstanceExtension(qry *query.Query, psh push.Pusher, log *slog.Logger) *SubstanceExtension {
	return &SubstanceExtension{
		qry: qry,
		psh: psh,
		log: log,
	}
}

type SubstanceExtension struct {
	qry *query.Query
	psh push.Pusher
	log *slog.Logger
}

func (se *SubstanceExtension) Create(ctx context.Context, req *request.SubstanceExtensionCreate, cu *session.Ident) error {
	now := time.Now()

	var brokID int64
	name, minionID := req.Name, req.MinionID
	tbl := se.qry.Substance
	dao := tbl.WithContext(ctx)
	if minionID != 0 {
		// 检查节点
		monTbl := se.qry.Minion
		mon, err := monTbl.WithContext(ctx).
			Select(monTbl.Status, monTbl.BrokerID, monTbl.Inet).
			Where(monTbl.ID.Eq(minionID)).
			First()
		if err != nil {
			return errcode.ErrNodeNotExist
		}
		status := mon.Status
		if status != model.MSOffline && status != model.MSOnline {
			return errcode.ErrNodeStatus
		}
		// 私有配置检查配置名是否存在
		if count, err := dao.Where(tbl.Name.Eq(name), tbl.MinionID.Eq(minionID)).
			Or(tbl.Name.Eq(name), tbl.MinionID.Eq(0)).
			Count(); count != 0 || err != nil {
			return errcode.FmtErrNameExist.Fmt(name)
		}

		brokID = mon.BrokerID
	} else { // 公有配置检查名字是否重复
		if count, err := dao.Where(tbl.Name.Eq(name)).
			Count(); err != nil || count != 0 {
			return errcode.FmtErrNameExist.Fmt(name)
		}
	}

	extensionID, data := req.ExtensionID, req.Data
	mktTbl := se.qry.ExtensionMarket
	mktDao := mktTbl.WithContext(ctx)
	extension, err := mktDao.Where(mktTbl.ID.Eq(extensionID), mktTbl.Category.Eq("service")).First()
	if err != nil {
		return err
	}

	tmpl, err := luatemplate.New(req.Name).Parse(extension.Content)
	if err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, req.Data); err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	}
	code := buf.Bytes()
	md5sum := md5.Sum(code)
	md5s := hex.EncodeToString(md5sum[:])

	dat := &model.Substance{
		Name:      req.Name,
		Hash:      md5s,
		Chunk:     code,
		MinionID:  minionID,
		Version:   1,
		CreatedID: cu.ID,
		UpdatedID: cu.ID,
		ContentQuote: &model.ExtensionQuote{
			ID:          extensionID,
			Name:        extension.Name,
			Intro:       extension.Intro,
			Version:     extension.Version,
			Data:        data,
			Content:     extension.Content,
			ContentSHA1: extension.ContentSHA1,
			CreatedBy:   extension.CreatedBy,
			UpdatedBy:   extension.UpdatedBy,
			CreatedAt:   extension.CreatedAt,
			UpdatedAt:   extension.UpdatedAt,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err = dao.Create(dat); err != nil {
		return err
	}
	if minionID != 0 && brokID != 0 { // 推送
		se.psh.TaskSync(ctx, brokID, []int64{minionID})
	}

	return nil
}

func (se *SubstanceExtension) Update(ctx context.Context, req *request.SubstanceExtensionUpdate, cu *session.Ident) error {
	subID, now := req.ID, time.Now()
	tbl := se.qry.Substance
	dao := tbl.WithContext(ctx)
	sub, err := dao.Where(tbl.ID.Eq(subID)).First()
	if err != nil {
		return err
	} else if sub.ContentQuote == nil {
		return errcode.ErrSubstanceNotExist
	}

	quote := sub.ContentQuote
	tmpl, err := luatemplate.New(sub.Name).Parse(quote.Content)
	if err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, req.Data); err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	}
	code := buf.Bytes()
	md5sum := md5.Sum(code)
	md5s := hex.EncodeToString(md5sum[:])
	quote.Data = req.Data

	if _, err = dao.Where(tbl.ID.Eq(subID)).
		UpdateSimple(
			tbl.UpdatedID.Value(cu.ID),
			tbl.UpdatedAt.Value(now),
			tbl.Chunk.Value(code),
			tbl.Hash.Value(md5s),
			tbl.ContentQuote.Value(quote),
		); err != nil {
		return err
	}
	if sub.MinionID == 0 {
		return nil
	}
	// 查询 agent

	return nil
}
