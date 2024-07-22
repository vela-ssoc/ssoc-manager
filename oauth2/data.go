package oauth2

import "time"

type Userinfo struct {
	SUB         string // 用户的咚咚号
	Name        string // 用户名
	Gender      bool   // 是否女性
	CompanyName string // 公司名称
	Picture     string // 头像
	Department  string // 部门全名
	ISS         string
	AUD         []string
	Scope       []string
	NBF         time.Time
	EXP         time.Time
	IAT         time.Time
	UpdatedAt   time.Time
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
	SUB                string   `json:"sub"`
	Gender             string   `json:"gender"`
	CompanyName        string   `json:"companyName"`
	ISS                string   `json:"iss"`
	Picture            string   `json:"picture"`
	AUD                []string `json:"aud"`
	NBF                float64  `json:"nbf"`
	UpdatedAt          string   `json:"updated_at"`
	Scope              []string `json:"scope"`
	Name               string   `json:"name"`
	DepartmentFullName string   `json:"departmentFullName"`
	EXP                float64  `json:"exp"`
	IAT                float64  `json:"iat"`
}

func (u dongUserinfoResponse) Userinfo() *Userinfo {
	updatedAt, _ := time.Parse(time.DateOnly, u.UpdatedAt)

	return &Userinfo{
		SUB:         u.SUB,
		Name:        u.Name,
		Gender:      u.Gender == "F",
		CompanyName: u.CompanyName,
		Picture:     u.Picture,
		Department:  u.DepartmentFullName,
		ISS:         u.ISS,
		AUD:         u.AUD,
		Scope:       u.Scope,
		NBF:         time.UnixMicro(int64(u.NBF)),
		EXP:         time.UnixMicro(int64(u.EXP)),
		IAT:         time.UnixMicro(int64(u.IAT)),
		UpdatedAt:   updatedAt,
	}
}
