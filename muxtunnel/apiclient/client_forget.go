package apiclient

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/httpclient"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
)

type ClientForget struct {
	hub muxproto.Huber
	cli httpclient.Client
	log *slog.Logger
}

func (cf ClientForget) HTTPProxy(ctx context.Context, location string) error {
	var errs []error
	for _, node := range cf.hub.All() {
		err := cf.httpProxy(ctx, node.Host(), location)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (cf ClientForget) Redis(ctx context.Context) error {
	var errs []error
	for _, node := range cf.hub.All() {
		err := cf.redis(ctx, node.Host())
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (cf ClientForget) VictoriaMetrics(ctx context.Context) error {
	var errs []error
	for _, node := range cf.hub.All() {
		err := cf.victoriaMetrics(ctx, node.FQDN())
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (cf ClientForget) httpProxy(parent context.Context, fqdn, location string) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()

	addr := muxproto.BuildFQDNURL(fqdn, "/api/v3/reverse-proxy")
	query := url.Values{"location": []string{location}}
	addr.RawQuery = query.Encode()
	strURL := addr.String()

	return cf.cli.DeleteJSON(ctx, strURL, nil, new(empty))
}

func (cf ClientForget) redis(parent context.Context, fqdn string) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()

	addr := muxproto.BuildFQDNURL(fqdn, "/api/v3/redis-config")
	strURL := addr.String()

	return cf.cli.DeleteJSON(ctx, strURL, nil, new(empty))
}

func (cf ClientForget) victoriaMetrics(parent context.Context, fqdn string) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()

	addr := muxproto.BuildFQDNURL(fqdn, "/api/v3/victoria-metrics")
	strURL := addr.String()

	return cf.cli.DeleteJSON(ctx, strURL, nil, new(empty))
}
