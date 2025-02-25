package confload

import (
	"context"
	"net/url"

	"github.com/vela-ssoc/vela-manager/integration/casauth"
	"github.com/vela-ssoc/vela-manager/integration/oauth"
)

func NewCASConfig(rawURL string) casauth.Configurer {
	u, e := url.Parse(rawURL)
	return &casConfig{
		e: e,
		u: u,
	}
}

type casConfig struct {
	e error
	u *url.URL
}

func (cc *casConfig) Configure(_ context.Context) (*url.URL, error) {
	if cc.e != nil {
		return nil, cc.e
	}

	cu := *cc.u // clone

	return &cu, nil
}

func NewOauthConfig(rawURL, clientID, clientSecret, redirectURL string) oauth.Configurer {
	oc := &oauthConfig{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
	}

	u, e := url.Parse(rawURL)
	if e != nil {
		oc.err = e
	} else {
		oc.destURL = u
	}

	return oc
}

type oauthConfig struct {
	destURL      *url.URL
	clientID     string
	clientSecret string
	redirectURL  string
	err          error
}

func (oc *oauthConfig) Configure(_ context.Context) (*oauth.Config, error) {
	if oc.err != nil {
		return nil, oc.err
	}

	destURL := *oc.destURL
	cfg := &oauth.Config{
		URL:          &destURL,
		ClientID:     oc.clientID,
		ClientSecret: oc.clientSecret,
		RedirectURL:  oc.redirectURL,
	}

	return cfg, nil
}
