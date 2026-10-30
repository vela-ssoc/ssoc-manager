package response

import "go.mongodb.org/mongo-driver/v2/bson"

type OAuthProvider struct {
	ClientID              string `json:"client_id"`
	RedirectURI           string `json:"redirect_uri"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

type AuthInfo struct {
	ID        bson.ObjectID `json:"id"`
	Company   string        `json:"company"` // 集团-group ,证券-security ,天天基金-ttfund ,期货-futures , 东财基金-dcfund
	Username  string        `json:"username"`
	Nickname  string        `json:"nickname"`
	Token     string        `json:"token"`
	AvatarURL string        `json:"avatar_url"`
	OAuth     bool          `json:"oauth"` // 是否通过 OAuth 登录
}
