package service

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/memcache"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type VictoriaMetrics struct {
	db    polydb.Sourcer
	rtp   http.RoundTripper
	log   *slog.Logger
	label string
	mem   *memcache.Value[*model.VictoriaMetricsConfig]
}

func NewVictoriaMetrics(db polydb.Sourcer, rtp http.RoundTripper, log *slog.Logger) *VictoriaMetrics {
	vmc := &VictoriaMetrics{
		db:    db,
		rtp:   rtp,
		log:   log,
		label: `appname="ssoc-manager"`, // 固定 Label
	}
	vmc.mem = memcache.NewValue(vmc.Get)

	return vmc
}

func (vmc *VictoriaMetrics) Get(ctx context.Context) (*model.VictoriaMetricsConfig, error) {
	mdb := vmc.db.MDB()
	coll := mdb.VictoriaMetricsConfig()

	return coll.FindOne(ctx, bson.D{})
}

func (vmc *VictoriaMetrics) Upsert(ctx context.Context, param request.VictoriaMetricsConfigUpsert) error {
	now := time.Now()
	mod := &model.VictoriaMetricsConfig{
		Enabled:   param.Enabled,
		Address:   param.Address,
		TenantID:  param.TenantID,
		Username:  param.Username,
		Password:  param.Password,
		Headers:   param.Headers.Canonical(),
		UpdatedAt: now,
	}
	mdb := vmc.db.MDB()
	coll := mdb.VictoriaMetricsConfig()
	opts := options.UpdateOne().SetUpsert(true)
	update := bson.M{"$set": mod, "$setOnInsert": bson.M{"created_at": now}}
	_, err := coll.UpdateOne(ctx, bson.D{}, update, opts)
	if err == nil {
		_, _ = vmc.mem.Forget()
	}

	return err
}

func (vmc *VictoriaMetrics) Reverse(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pu, headers, data, err := vmc.load(ctx)
	if err != nil {
		return err
	}

	pre := data.SelectPrefixPath()
	pu = pu.JoinPath(pre)

	pxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
			pr.SetURL(pu)
			for k, v := range headers {
				pr.Out.Header.Set(k, v)
			}
		},
		Transport: vmc.rtp,
		ModifyResponse: func(r *http.Response) error {
			if r.StatusCode == http.StatusUnauthorized {
				r.StatusCode = http.StatusBadRequest
				vmc.log.Warn("VictoriaMetrics 用户名密码错误")
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			vmc.log.Warn("VictoriaMetrics 请求响应错误", "err", err)
		},
	}
	pxy.ServeHTTP(w, r)

	return nil
}

func (vmc *VictoriaMetrics) LoadVictoriaPushConfig(ctx context.Context) (pushURL string, opts *metrics.PushOptions, err error) {
	pu, headers, data, err := vmc.load(ctx)
	if err != nil {
		return "", nil, err
	}

	pre := data.InsertPrefixPath()
	suffix := path.Join(pre, "/api/v1/import/prometheus")
	pu = pu.JoinPath(suffix)
	opts = &metrics.PushOptions{ExtraLabels: vmc.label, Headers: headers.Array()}

	return pu.String(), opts, nil
}

func (vmc *VictoriaMetrics) load(ctx context.Context) (*url.URL, model.HTTPHeader, *model.VictoriaMetricsConfig, error) {
	dat, err := vmc.mem.Load(ctx)
	if err != nil {
		return nil, nil, nil, errcode.ErrVictoriaNotConfigured
	}
	if !dat.Enabled {
		return nil, nil, nil, errcode.ErrVictoriaDisabled
	}
	addr, header, err := dat.Target()
	if err != nil {
		vmc.log.Error("VictoriaMetrics 配置参数错误", "config", dat, "err", err)
		return nil, nil, nil, err
	}

	return addr, header, dat, nil
}
