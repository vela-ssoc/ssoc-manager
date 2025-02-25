package oauth

import "time"

type Userinfo struct {
	SUB         string // 用户的咚咚号
	Name        string // 用户名
	Gender      bool   // 是否女性
	CompanyName string // 公司名称
	Picture     string // 头像
	Department  string // 部门全名
	ISS         string //
}

type QrcodeStatus struct {
	SiteBase    string
	UUID        string
	ClientName  string
	ExpireAt    time.Time
	RedirectURI string
	Status      string
	Token       string
	TicketID    string
	Code        string
}

type dongResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (dr *dongResponse) Error() string {
	return dr.Code + dr.Msg
}

func (dr *dongResponse) AutoError() error {
	if dr.Code == "200" {
		return nil
	}
	return dr
}

type qrcodeStatus struct {
	SiteBase    string `json:"siteBase"`
	UUID        string `json:"uuid"`
	ClientName  string `json:"clientName"`
	ExpireAt    int64  `json:"expireAt"`
	RedirectURI string `json:"redirectUri"`
	Status      string `json:"status"`
	Token       string `json:"token"`
	TicketID    string `json:"ticketId"`
	Code        string `json:"code"`
}

func (qs qrcodeStatus) QrcodeStatus() *QrcodeStatus {
	return &QrcodeStatus{
		SiteBase:    qs.SiteBase,
		UUID:        qs.UUID,
		ClientName:  qs.ClientName,
		ExpireAt:    time.UnixMicro(qs.ExpireAt),
		RedirectURI: qs.RedirectURI,
		Status:      qs.Status,
		Token:       qs.Token,
		TicketID:    qs.TicketID,
		Code:        qs.Code,
	}
}

type connectRequest struct {
	ClientID     string `json:"clientId"`
	Scope        string `json:"scope"`
	RedirectURI  string `json:"redirectUri"`
	ResponseType string `json:"responseType"`
	State        string `json:"state"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (tr tokenResponse) Token() string {
	return tr.TokenType + " " + tr.AccessToken
}

type dongUserinfoResponse struct {
	SUB                string `json:"sub"`                // 工号（不带前缀，即证券用户为"17290"而非"Z17290"）
	CompanyPrefix      string `json:"companyPrefix"`      // 公司：G-集团，Z-证券，T-天天基金
	Gender             string `json:"gender"`             // 性别：F-女，M-男
	CompanyName        string `json:"companyName"`        // 公司名称
	UniqueKey          string `json:"uniqueKey"`          // 唯一ID
	Nonce              string `json:"nonce"`              // 跳转登录系统前端时如果携带了 nonce，这里就会原样传回
	Picture            string `json:"picture"`            // 头像
	Name               string `json:"name"`               // 员工姓名
	DepartmentFullName string `json:"departmentFullName"` // 部门
	Email              string `json:"email"`              // 邮箱
}

func (u dongUserinfoResponse) Userinfo() *Userinfo {
	return &Userinfo{
		SUB:         u.SUB,
		Name:        u.Name,
		Gender:      u.Gender == "F",
		CompanyName: u.CompanyName,
		Picture:     u.Picture,
		Department:  u.DepartmentFullName,
	}
}
