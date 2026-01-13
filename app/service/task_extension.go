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
	"github.com/vela-ssoc/ssoc-common-mb/cronv3"
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-common-mb/param/response"
	"github.com/vela-ssoc/ssoc-manager/app/service/internal/minionfilter"
	"github.com/vela-ssoc/ssoc-manager/app/service/internal/taskexec"
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"gorm.io/gen/field"
)

func NewTaskExtension(qry *query.Query, hub linkhub.Huber, flt *minionfilter.Filter, crontab *cronv3.Crontab) *TaskExtension {
	exec := taskexec.New(qry, hub, flt)
	return &TaskExtension{
		qry:     qry,
		crontab: crontab,
		exec:    exec,
	}
}

type TaskExtension struct {
	qry     *query.Query
	crontab *cronv3.Crontab // 定时器
	exec    *taskexec.TaskExec
}

func (tim *TaskExtension) Init(ctx context.Context) {
	tbl := tim.qry.TaskExtension
	dao := tbl.WithContext(ctx)

	ors := dao.Where(tbl.Cron.Neq("")).Or(tbl.SpecificTimes.IsNotNull())
	tasks, _ := dao.Where(tbl.Enabled.Is(true), ors).Find()

	for _, task := range tasks {
		cronName := tim.taskName(task.ID)
		cronFunc := tim.execute(task.ID)
		if len(task.SpecificTimes) != 0 {
			times := cronv3.NewSpecificTimes(task.SpecificTimes)
			tim.crontab.Schedule(cronName, times, cronFunc)
		} else if task.Cron != "" {
			tim.crontab.AddFunc(cronName, task.Cron, cronFunc)
		}
	}
}

func (tim *TaskExtension) FromMarket(ctx context.Context, req *mrequest.TaskExtensionFromMarket, cu *session.Ident) error {
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
		Intro:       market.Intro,
		Version:     market.Version,
		Data:        req.Data,
		Content:     market.Content,
		ContentSHA1: market.ContentSHA1,
		CreatedBy:   market.CreatedBy,
		UpdatedBy:   market.UpdatedBy,
		CreatedAt:   market.CreatedAt,
		UpdatedAt:   market.UpdatedAt,
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

func (tim *TaskExtension) FromCode(ctx context.Context, req *mrequest.TaskExtensionFromCode, cu *session.Ident) error {
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
	tbl := tim.qry.TaskExtension
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Delete()

	return err
}

func (tim *TaskExtension) Page(ctx context.Context, req *request.PageKeywords) (*response.Pages[*model.TaskExtension], error) {
	tbl := tim.qry.TaskExtension
	dao := tbl.WithContext(ctx)

	if kw := req.Keyword; kw != "" {
		kw = "%" + kw + "%"
		dao = dao.Where(tbl.Name.Like(kw)).Or(tbl.Intro.Like(kw))
	}

	page := response.NewPages[*model.TaskExtension](req.PageSize())
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return page.Empty(), nil
	}

	records, err := dao.Order(tbl.ID.Desc()).
		Scopes(page.FP(cnt)).
		Find()
	if err != nil {
		return nil, err
	}

	return page.SetRecords(records), nil
}

func (tim *TaskExtension) CreateCode(ctx context.Context, req *mrequest.TaskExtensionCreateCode, cu *session.Ident) (*model.TaskExtension, error) {
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
			Intro:       mrk.Intro,
			Version:     mrk.Version,
			Data:        req.Data,
			Content:     mrk.Content,
			ContentSHA1: mrk.ContentSHA1,
			CreatedBy:   mrk.CreatedBy,
			UpdatedBy:   mrk.UpdatedBy,
			CreatedAt:   mrk.CreatedAt,
			UpdatedAt:   mrk.UpdatedAt,
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

func (tim *TaskExtension) UpdateCode(ctx context.Context, req *mrequest.TaskExtensionUpdateCode, cu *session.Ident) (*model.TaskExtension, error) {
	tbl := tim.qry.TaskExtension
	dao := tbl.WithContext(ctx)
	// 查询已存在的数据
	old, err := dao.Where(tbl.ID.Eq(req.ID)).First()
	if err != nil {
		return nil, err
	}

	code := req.Code
	operator := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}
	updates := []field.AssignExpr{
		tbl.Intro.Value(req.Intro),
		tbl.UpdatedBy.Value(operator),
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
			Intro:       mrk.Intro,
			Version:     mrk.Version,
			Data:        req.Data,
			Content:     mrk.Content,
			ContentSHA1: mrk.ContentSHA1,
			CreatedBy:   mrk.CreatedBy,
			UpdatedBy:   mrk.UpdatedBy,
			CreatedAt:   mrk.CreatedAt,
			UpdatedAt:   mrk.UpdatedAt,
		}
		updates = append(updates, tbl.ContentQuote.Value(contentQuote))

		name, content := mrk.Name, mrk.Content
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

func (tim *TaskExtension) CreatePublish(ctx context.Context, req *mrequest.TaskExtensionCreatePublish, cu *session.Ident) error {
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
		Filters:   req.Filters.ConvertModel(),
		Excludes:  req.Excludes,
		CreatedBy: operator,
		UpdatedBy: operator,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if cron := req.Cron; cron != "" {
		data.Cron = cron
	} else if sts := req.SpecificTimes; len(sts) != 0 {
		data.SpecificTimes = sts
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
			Intro:       mrk.Intro,
			Version:     mrk.Version,
			Data:        req.Data,
			Content:     mrk.Content,
			ContentSHA1: mrk.ContentSHA1,
			CreatedBy:   mrk.CreatedBy,
			UpdatedBy:   mrk.UpdatedBy,
			CreatedAt:   mrk.CreatedAt,
			UpdatedAt:   mrk.UpdatedAt,
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

	taskExtensionDo := tim.qry.TaskExtension.WithContext(ctx)
	if err = taskExtensionDo.Create(data); err != nil || !enabled {
		return err
	}

	cronName := tim.taskName(data.ID)
	if cron := req.Cron; cron != "" {
		tim.crontab.AddFunc(cronName, cron, tim.execute(data.ID))
	} else if sts := req.SpecificTimes; len(sts) != 0 {
		times := cronv3.NewSpecificTimes(sts)
		tim.crontab.Schedule(cronName, times, tim.execute(data.ID))
	} else {
		go tim.execute(data.ID)()
	}

	return nil
}

func (tim *TaskExtension) UpdatePublish(ctx context.Context, req *mrequest.TaskExtensionUpdatePublish, cu *session.Ident) error {
	code, enabled := req.Code, req.Enabled
	operator := model.Operator{ID: cu.ID, Nickname: cu.Nickname, Username: cu.Username}

	tbl := tim.qry.TaskExtension
	dao := tbl.WithContext(ctx)
	old, err := dao.Where(tbl.ID.Eq(req.ID)).First()
	if err != nil {
		return err
	}
	// 立即执行的任务，运行中不允许修改。
	if !old.Finished && old.Status != nil && old.Cron == "" && len(old.SpecificTimes) == 0 {
		return errcode.ErrEditRunningTask
	}

	updates := []field.AssignExpr{
		tbl.Intro.Value(req.Intro),
		tbl.Enabled.Value(enabled),
		tbl.PushSize.Value(req.PushSize),
		tbl.UpdatedBy.Value(operator),
		tbl.Filters.Value(req.Filters.ConvertModel()),
		tbl.Excludes.Value(req.Excludes),
		tbl.StepDone.Value(true),
		tbl.Timeout.Value(req.Timeout),
	}
	if cron := req.Cron; cron != "" {
		updates = append(updates, tbl.Cron.Value(cron), tbl.SpecificTimes.Value(nil))
	} else if sts := req.SpecificTimes; len(sts) != 0 {
		updates = append(updates, tbl.SpecificTimes.Value(sts), tbl.Cron.Value(""))
	} else {
		updates = append(updates, tbl.Cron.Value(""), tbl.SpecificTimes.Value(nil))
	}

	if old.Code != "" && code != "" {
		sum := sha1.Sum([]byte(code))
		codeSHA1 := hex.EncodeToString(sum[:])
		updates = append(updates, tbl.Code.Value(code), tbl.CodeSHA1.Value(codeSHA1))
	} else if quote := old.ContentQuote; quote != nil {
		extensionID := req.ExtensionID
		mrkTbl := tim.qry.ExtensionMarket
		mrk, err1 := mrkTbl.WithContext(ctx).
			Where(mrkTbl.ID.Eq(extensionID), mrkTbl.Category.Eq("task")).
			First()
		if err1 != nil {
			return err1
		}
		contentQuote := &model.ExtensionQuote{
			ID:          mrk.ID,
			Name:        mrk.Name,
			Intro:       mrk.Intro,
			Version:     mrk.Version,
			Data:        req.Data,
			Content:     mrk.Content,
			ContentSHA1: mrk.ContentSHA1,
			CreatedBy:   mrk.CreatedBy,
			UpdatedBy:   mrk.UpdatedBy,
			CreatedAt:   mrk.CreatedAt,
			UpdatedAt:   mrk.UpdatedAt,
		}
		updates = append(updates, tbl.ContentQuote.Value(contentQuote))

		name, content := mrk.Name, mrk.Content
		tmpl, err2 := luatemplate.New(name).Parse(content)
		if err2 != nil {
			return errcode.FmtErrGenerateCode.Fmt(err2)
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

	if _, err = tim.qry.TaskExtension.
		WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		UpdateSimple(updates...); err != nil || !req.Enabled {
		return err
	}

	cronName := tim.taskName(req.ID)
	if cron := req.Cron; cron != "" {
		tim.crontab.AddFunc(cronName, cron, tim.execute(req.ID))
	} else if sts := req.SpecificTimes; len(sts) != 0 {
		times := cronv3.NewSpecificTimes(sts)
		tim.crontab.Schedule(cronName, times, tim.execute(req.ID))
	} else {
		go tim.execute(req.ID)()
	}

	return nil
}

func (tim *TaskExtension) taskName(id int64) string {
	return "task-extension:" + strconv.FormatInt(id, 10)
}

func (tim *TaskExtension) execute(id int64) func() {
	return func() {
		_ = tim.exec.Exec(context.Background(), id)
	}
}
