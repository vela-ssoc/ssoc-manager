package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/vela-public/onekit/lua/parse"
	"github.com/vela-ssoc/luatemplate"
	"github.com/vela-ssoc/vela-common-mb/cronv3"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/gen/field"
)

func NewTaskExtension(qry *query.Query, crontab *cronv3.Crontab) *TaskExtension {
	return &TaskExtension{
		qry:     qry,
		crontab: crontab,
	}
}

type TaskExtension struct {
	qry     *query.Query
	crontab *cronv3.Crontab // 定时器
}

func (tim *TaskExtension) FromMarket(ctx context.Context, req *param.TaskExtensionFromMarket, cu *session.Ident) error {
	eid, now := req.ExtensionID, time.Now()
	mktTbl := tim.qry.ExtensionMarket
	market, err := mktTbl.WithContext(ctx).
		Where(mktTbl.ID.Eq(eid), mktTbl.Category.Eq("task")).
		First()
	if err != nil {
		return err
	}

	tmpl, err := luatemplate.New(market.Name).Parse(market.Content)
	if err != nil {
		return err
	}

	h1sum := sha1.New()
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(io.MultiWriter(h1sum, buf), req.Data); err != nil {
		return err
	}
	sum := h1sum.Sum(nil)
	codeSHA1 := hex.EncodeToString(sum)

	code := buf.String()
	stmts, err := parse.Parse(buf, req.Name)
	if err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	} else if len(stmts) == 0 {
		return errcode.ErrGenerateEmptyCode
	}

	createdBy := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}
	quote := &model.ExtensionQuote{
		ID:          market.ID,
		Name:        market.Name,
		Version:     market.Version,
		Data:        req.Data,
		Content:     market.Content,
		ContentSHA1: market.ContentSHA1,
		CreatedBy:   market.CreatedBy,
		UpdatedBy:   market.UpdatedBy,
	}
	data := &model.TaskExtension{
		Name:         req.Name,
		Intro:        req.Intro,
		Code:         code,
		CodeSHA1:     codeSHA1,
		ContentQuote: quote,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return tim.qry.TaskExtension.WithContext(ctx).Create(data)
}

func (tim *TaskExtension) FromCode(ctx context.Context, req *param.TaskExtensionFromCode, cu *session.Ident) error {
	now, code := time.Now(), req.Code
	sum := sha1.Sum([]byte(code))
	codeSHA1 := hex.EncodeToString(sum[:])

	stmts, err := parse.Parse(strings.NewReader(code), req.Name)
	if err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	} else if len(stmts) == 0 {
		return errcode.ErrGenerateEmptyCode
	}

	createdBy := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}
	data := &model.TaskExtension{
		Name:      req.Name,
		Intro:     req.Intro,
		Code:      code,
		CodeSHA1:  codeSHA1,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return tim.qry.TaskExtension.WithContext(ctx).Create(data)
}

func (tim *TaskExtension) Delete(ctx context.Context, id int64) error {
	return nil
}

func (tim *TaskExtension) Page(ctx context.Context, page param.Pager) (int64, []*model.TaskExtension) {
	tbl := tim.qry.TaskExtension
	dao := tbl.WithContext(ctx)
	if kw := page.Keyword(); kw != "" {
		dao = dao.Where(tbl.Name.Like(kw)).Or(tbl.Intro.Like(kw))
	}

	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (tim *TaskExtension) Release(ctx context.Context, req *param.TaskExtensionRelease) error {
	//tbl := tim.qry.TaskExtension
	//dao := tbl.WithContext(ctx)
	//taskExt, err := dao.Where(tbl.ID.Eq(req.ID)).First()
	//if err != nil {
	//	return err
	//}

	return nil
}

func (tim *TaskExtension) CreateRelease(ctx context.Context) error {
	return nil
}

func (tim *TaskExtension) UpdateRelease(ctx context.Context) error {
	return nil
}

func (tim *TaskExtension) CreateCode(ctx context.Context, req *param.TaskExtensionCreateCode, cu *session.Ident) (*model.TaskExtension, error) {
	code, now := req.Code, time.Now()
	operator := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}
	data := &model.TaskExtension{
		Name:      req.Name,
		Intro:     req.Intro,
		CreatedBy: operator,
		UpdatedBy: operator,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if code != "" {
		data.Code = code
		sum := sha1.Sum([]byte(code))
		data.CodeSHA1 = hex.EncodeToString(sum[:])
	} else {
		tbl := tim.qry.ExtensionMarket
		mrk, err := tbl.WithContext(ctx).
			Where(tbl.ID.Eq(req.ExtensionID), tbl.Category.Eq("task")).
			First()
		if err != nil {
			return nil, err
		}
		data.ContentQuote = &model.ExtensionQuote{
			ID:          mrk.ID,
			Name:        mrk.Name,
			Version:     mrk.Version,
			Data:        req.Data,
			Content:     mrk.Content,
			ContentSHA1: mrk.ContentSHA1,
			CreatedBy:   mrk.CreatedBy,
			UpdatedBy:   mrk.UpdatedBy,
		}
		tmpl, err := luatemplate.New(mrk.Name).Parse(mrk.Content)
		if err != nil {
			return nil, errcode.FmtErrGenerateCode.Fmt(err)
		}
		buf := new(bytes.Buffer)
		if err = tmpl.Execute(buf, req.Data); err != nil {
			return nil, errcode.FmtErrGenerateCode.Fmt(err)
		}
		sum := sha1.Sum(buf.Bytes())
		data.CodeSHA1 = hex.EncodeToString(sum[:])
		data.Code = buf.String()
	}

	tbl := tim.qry.TaskExtension
	err := tbl.WithContext(ctx).Create(data)

	return data, err
}

func (tim *TaskExtension) UpdateCode(ctx context.Context, req *param.TaskExtensionUpdateCode, cu *session.Ident) (*model.TaskExtension, error) {
	tbl := tim.qry.TaskExtension
	dao := tbl.WithContext(ctx)
	// 查询已存在的数据
	old, err := dao.Where(tbl.ID.Eq(req.ID)).First()
	if err != nil {
		return nil, err
	}

	code, now := req.Code, time.Now()
	operator := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}
	updates := []field.AssignExpr{
		tbl.Intro.Value(req.Intro),
		tbl.UpdatedBy.Value(operator),
		tbl.UpdatedAt.Value(now),
	}
	if old.Code != "" && code != "" {
		sum := sha1.Sum([]byte(code))
		codeSHA1 := hex.EncodeToString(sum[:])
		updates = append(updates,
			tbl.Code.Value(code),
			tbl.CodeSHA1.Value(codeSHA1),
		)
	} else if quote := old.ContentQuote; quote != nil {
		extensionID := req.ExtensionID
		var name, content string
		if extensionID != 0 && extensionID != quote.ID {
			mrkTbl := tim.qry.ExtensionMarket
			mrk, err := mrkTbl.WithContext(ctx).
				Where(mrkTbl.ID.Eq(extensionID), mrkTbl.Category.Eq("task")).
				First()
			if err != nil {
				return nil, err
			}
			contentQuote := &model.ExtensionQuote{
				ID:          mrk.ID,
				Name:        mrk.Name,
				Version:     mrk.Version,
				Data:        req.Data,
				Content:     mrk.Content,
				ContentSHA1: mrk.ContentSHA1,
				CreatedBy:   mrk.CreatedBy,
				UpdatedBy:   mrk.UpdatedBy,
			}
			updates = append(updates, tbl.ContentQuote.Value(contentQuote))
		} else {
			name = quote.Name
			content = quote.Content
		}

		tmpl, err := luatemplate.New(name).Parse(content)
		if err != nil {
			return nil, errcode.FmtErrGenerateCode.Fmt(err)
		}
		buf := new(bytes.Buffer)
		if err = tmpl.Execute(buf, req.Data); err != nil {
			return nil, errcode.FmtErrGenerateCode.Fmt(err)
		}
		code = buf.String()
		sum := sha1.Sum([]byte(code))
		codeSHA1 := hex.EncodeToString(sum[:])

		updates = append(updates,
			tbl.Code.Value(code),
			tbl.CodeSHA1.Value(codeSHA1),
		)
	}
	// 检测 Lua Code 是否合法
	stmts, err := parse.Parse(strings.NewReader(code), "")
	if err != nil {
		return nil, errcode.FmtErrGenerateCode.Fmt(err)
	} else if len(stmts) == 0 {
		return nil, errcode.ErrGenerateEmptyCode
	}

	if _, err = dao.Where(tbl.ID.Eq(req.ID)).UpdateSimple(updates...); err != nil {
		return nil, err
	}

	return dao.Where(tbl.ID.Eq(req.ID)).First()
}

func (tim *TaskExtension) CreatePublish(ctx context.Context, req *param.TaskExtensionCreatePublish, cu *session.Ident) error {
	now := time.Now()
	code, enabled := req.Code, req.Enabled
	operator := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}

	data := &model.TaskExtension{
		Name:      req.Name,
		Intro:     req.Intro,
		StepDone:  true,
		Enabled:   enabled,
		Timeout:   req.Timeout,
		PushSize:  req.PushSize,
		Filters:   req.Filters,
		Excludes:  req.Excludes,
		CreatedBy: operator,
		UpdatedBy: operator,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if code != "" {
		sum := sha1.Sum([]byte(code))
		codeSHA1 := hex.EncodeToString(sum[:])
		data.Code = code
		data.CodeSHA1 = codeSHA1
	} else {
		extensionID := req.ExtensionID
		mrkTbl := tim.qry.ExtensionMarket
		mrk, err := mrkTbl.WithContext(ctx).
			Where(mrkTbl.ID.Eq(extensionID), mrkTbl.Category.Eq("task")).
			First()
		if err != nil {
			return err
		}

		data.ContentQuote = &model.ExtensionQuote{
			ID:          mrk.ID,
			Name:        mrk.Name,
			Version:     mrk.Version,
			Data:        req.Data,
			Content:     mrk.Content,
			ContentSHA1: mrk.ContentSHA1,
			CreatedBy:   mrk.CreatedBy,
			UpdatedBy:   mrk.UpdatedBy,
		}

		tmpl, err := luatemplate.New(mrk.Name).Parse(mrk.Content)
		if err != nil {
			return errcode.FmtErrGenerateCode.Fmt(err)
		}
		buf := new(bytes.Buffer)
		if err = tmpl.Execute(buf, req.Data); err != nil {
			return errcode.FmtErrGenerateCode.Fmt(err)
		}

		code = buf.String()
		data.Code = code
		sum := sha1.Sum([]byte(code))
		data.CodeSHA1 = hex.EncodeToString(sum[:])
	}
	// 检测 Lua Code 是否合法
	stmts, err := parse.Parse(strings.NewReader(code), "")
	if err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	} else if len(stmts) == 0 {
		return errcode.ErrGenerateEmptyCode
	}

	return tim.qry.TaskExtension.WithContext(ctx).Create(data)
}

func (tim *TaskExtension) UpdatePublish(ctx context.Context, req *param.TaskExtensionUpdatePublish, cu *session.Ident) error {
	now := time.Now()
	code, enabled := req.Code, req.Enabled
	operator := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}

	tbl := tim.qry.TaskExtension
	dao := tbl.WithContext(ctx)
	old, err := dao.Where(tbl.ID.Eq(req.ID)).First()
	if err != nil {
		return err
	}

	updates := []field.AssignExpr{
		tbl.Intro.Value(req.Intro),
		tbl.Enabled.Value(enabled),
		tbl.PushSize.Value(req.PushSize),
		tbl.UpdatedBy.Value(operator),
		tbl.UpdatedAt.Value(now),
		tbl.Filters.Value(req.Filters),
		tbl.Excludes.Value(req.Excludes),
	}

	if old.Code != "" && code != "" {
		sum := sha1.Sum([]byte(code))
		codeSHA1 := hex.EncodeToString(sum[:])
		updates = append(updates, tbl.Code.Value(code), tbl.CodeSHA1.Value(codeSHA1))
	} else if quote := old.ContentQuote; quote != nil {
		extensionID := req.ExtensionID
		var name, content string
		if extensionID != 0 && extensionID != quote.ID {
			mrkTbl := tim.qry.ExtensionMarket
			mrk, err := mrkTbl.WithContext(ctx).
				Where(mrkTbl.ID.Eq(extensionID), mrkTbl.Category.Eq("task")).
				First()
			if err != nil {
				return err
			}
			contentQuote := &model.ExtensionQuote{
				ID:          mrk.ID,
				Name:        mrk.Name,
				Version:     mrk.Version,
				Data:        req.Data,
				Content:     mrk.Content,
				ContentSHA1: mrk.ContentSHA1,
				CreatedBy:   mrk.CreatedBy,
				UpdatedBy:   mrk.UpdatedBy,
			}
			updates = append(updates, tbl.ContentQuote.Value(contentQuote))
		} else {
			name = quote.Name
			content = quote.Content
		}

		tmpl, err := luatemplate.New(name).Parse(content)
		if err != nil {
			return errcode.FmtErrGenerateCode.Fmt(err)
		}
		buf := new(bytes.Buffer)
		if err = tmpl.Execute(buf, req.Data); err != nil {
			return errcode.FmtErrGenerateCode.Fmt(err)
		}
		code = buf.String()
		sum := sha1.Sum([]byte(code))
		codeSHA1 := hex.EncodeToString(sum[:])

		updates = append(updates,
			tbl.Code.Value(code),
			tbl.CodeSHA1.Value(codeSHA1),
		)
	}
	// 检测 Lua Code 是否合法
	stmts, err := parse.Parse(strings.NewReader(code), "")
	if err != nil {
		return errcode.FmtErrGenerateCode.Fmt(err)
	} else if len(stmts) == 0 {
		return errcode.ErrGenerateEmptyCode
	}

	_, err = tim.qry.TaskExtension.
		WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		UpdateSimple(updates...)

	return err
}

func (tim *TaskExtension) taskName(id int64) string {
	return "task-extension:" + strconv.FormatInt(id, 10)
}
