package request

type OAuthConfigUpsert struct {
	Enabled               bool     `json:"enabled"`
	ClientID              string   `json:"client_id"              validate:"required,lte=100"`
	ClientSecret          string   `json:"client_secret"          validate:"required,lte=100"`
	AuthorizationEndpoint string   `json:"authorization_endpoint" validate:"required,http_url"`
	TokenEndpoint         string   `json:"token_endpoint"         validate:"required,http_url"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"      validate:"required,http_url"`
	RedirectURIs          []string `json:"redirect_uris"          validate:"gte=1,lte=100,dive,required,http_url"`
	Scopes                []string `json:"scopes,omitzero"        validate:"lte=10,dive,required,lte=10"`
}

type OAuthDongLogin struct {
	Code        string `json:"code"         form:"code"         query:"code"         validate:"required,lte=1000"`
	RedirectURI string `json:"redirect_uri" form:"redirect_uri" query:"redirect_uri"`
}

type OAuthProvider struct {
	Origin string `json:"origin" query:"origin"`
}
