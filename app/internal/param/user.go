package param

type UserSummaries []*userSummary

type userSummary struct {
	ID       int64  `json:"id,string" gorm:"column:id"`
	Username string `json:"username"  gorm:"column:username"`
	Nickname string `json:"nickname"  gorm:"column:nickname"`
	Dong     string `json:"dong"      gorm:"column:dong"`
	Enable   bool   `json:"enable"    gorm:"column:enable"`
}

type UserSudo struct {
	IntID
	Nickname string `json:"nickname"  validate:"required,lte=20"`
	Enable   bool   `json:"enable"`
	Password string `json:"password"  validate:"omitempty,password"`
}
