package response

type OAuthProvider struct {
	ClientID              string `json:"client_id"`
	RedirectURI           string `json:"redirect_uri"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}
