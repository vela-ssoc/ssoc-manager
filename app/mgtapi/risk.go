package mgtapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func Risk(qry *query.Query, svc service.RiskService) route.Router {
	riskTypeCol := dynsql.StringColumn("risk_type", "风险类型").Build()
	subjectCol := dynsql.StringColumn("subject", "主题").Build()
	inetCol := dynsql.StringColumn("inet", "终端 IP").Build()
	levelCol := dynsql.StringColumn("level", "级别").Build()
	statusCol := dynsql.StringColumn("status", "状态").Build()
	fromCodeCol := dynsql.StringColumn("from_code", "来源模块").Build()

	table := dynsql.Builder().
		Filters(subjectCol, riskTypeCol, inetCol, levelCol, statusCol, fromCodeCol,
			dynsql.TimeColumn("occur_at", "产生时间").Build(),
			dynsql.StringColumn("region", "归属地").Build(),
			dynsql.StringColumn("remote_ip", "外部IP").Build(),
			dynsql.IntColumn("remote_port", "外部端口").Build(),
			dynsql.StringColumn("local_ip", "本地IP").Build(),
			dynsql.IntColumn("local_port", "本地端口").Build(),
			dynsql.StringColumn("payload", "攻击载荷").Build(),
			dynsql.IntColumn("minion_id", "终端ID").Build(),
			dynsql.IntColumn("id", "风险ID").Build(),
		).
		Groups(subjectCol, riskTypeCol, inetCol, levelCol, statusCol, fromCodeCol).
		Build()
	return &riskREST{
		svc:   svc,
		table: table,
	}
}

type riskREST struct {
	qry   *query.Query
	svc   service.RiskService
	table dynsql.Table
}

func (rest *riskREST) Route(anon, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/risk/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/risk/attack").Data(route.Ignore()).GET(rest.Attack)
	anon.Route("/risk/group").Data(route.Ignore()).GET(rest.Group)
	anon.Route("/risk/recent").Data(route.Ignore()).GET(rest.Recent)
	anon.Route("/risks").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/risk/csv").Data(route.Ignore()).GET(rest.CSV)
	anon.Route("/risk/pie").Data(route.Ignore()).GET(rest.Pie)
	bearer.Route("/risk").
		Data(route.Named("批量删除风险事件")).DELETE(rest.Delete)
	anon.Route("/risk").Data(route.Ignore()).GET(rest.HTML)
	anon.Route("/risk/payloads").Data(route.Ignore()).GET(rest.Payloads)
	bearer.Route("/risk/ignore").
		Data(route.Named("批量忽略风险事件")).PATCH(rest.Ignore)
	bearer.Route("/risk/process").
		Data(route.Named("批量处理风险事件")).PATCH(rest.Process)
}

func (rest *riskREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *riskREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *riskREST) Attack(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Attack(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *riskREST) Group(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	if req.Group == "" {
		return errcode.ErrRequiredGroup
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Group(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *riskREST) Recent(c *ship.Context) error {
	day, _ := strconv.Atoi(c.Query("day"))
	if day > 30 || day < 1 { // 最多支持30天内查询，参数错误或超过有效范围默认为7天
		day = 7
	}

	ctx := c.Request().Context()
	res := rest.svc.Recent(ctx, day)

	return c.JSON(http.StatusOK, res)
}

func (rest *riskREST) CSV(c *ship.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func (rest *riskREST) Pie(c *ship.Context) error {
	group := c.Query("group")
	rtype := c.Query("risk_type")
	topN, _ := strconv.Atoi(c.Query("topn"))
	if topN <= 0 || topN >= 100 {
		topN = 10
	}

	ctx := c.Request().Context()
	tx := rest.qry.Risk.WithContext(ctx).UnderlyingDB()
	if rtype != "" {
		tx = tx.Where("risk_type = ?", rtype)
	}

	res := &mrequest.PieTopN{TopN: make(request.NameCounts, 0, topN)}
	var count int64
	if tx.Count(&count); count == 0 {
		return c.JSON(http.StatusOK, res)
	}
	tx.Select("COUNT(*) count", group+" name").
		Group(group).
		Order("count DESC").
		Limit(topN).
		Scan(&res.TopN)

	var num int
	for _, tn := range res.TopN {
		num += tn.Count
	}
	res.Other = int(count) - num

	return c.JSON(http.StatusOK, res)
}

func (rest *riskREST) Delete(c *ship.Context) error {
	var req dynsql.Input
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.Filters) == 0 {
		return errcode.ErrRequiredFilter
	}
	scope, err := rest.table.Inter(req)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, scope)
}

func (rest *riskREST) Ignore(c *ship.Context) error {
	var req dynsql.Input
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.Filters) == 0 {
		return errcode.ErrRequiredFilter
	}
	scope, err := rest.table.Inter(req)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	ctx := c.Request().Context()

	return rest.svc.Ignore(ctx, scope)
}

func (rest *riskREST) Process(c *ship.Context) error {
	var req dynsql.Input
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.Filters) == 0 {
		return errcode.ErrRequiredFilter
	}
	scope, err := rest.table.Inter(req)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	ctx := c.Request().Context()

	return rest.svc.Process(ctx, scope)
}

func (rest *riskREST) HTML(c *ship.Context) error {
	var req mrequest.ViewHTML
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	buf := rest.svc.HTML(ctx, req.ID, req.Secret)
	size := strconv.FormatInt(int64(buf.Len()), 10)
	c.SetRespHeader(ship.HeaderContentLength, size)

	return c.Stream(http.StatusOK, ship.MIMETextHTMLCharsetUTF8, buf)
}

func (rest *riskREST) Payloads(c *ship.Context) error {
	var req mrequest.RiskPayloadRequest
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	days := req.Days
	if days <= 0 {
		days = 1
	}
	now := time.Now()
	start := now.Add(-time.Duration(days) * time.Hour * 24)

	ret, err := rest.svc.Payloads(ctx, req.Pages, start, now, req.RiskType)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
