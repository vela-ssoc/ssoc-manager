package param

type DashGoosVO struct {
	Linux   int `json:"linux"`
	Windows int `json:"windows"`
	Darwin  int `json:"darwin"`
}

type DashStatusResp struct {
	Online   int `json:"online"   gorm:"column:online"`
	Offline  int `json:"offline"  gorm:"column:offline"`
	Inactive int `json:"inactive" gorm:"column:inactive"`
	Deleted  int `json:"deleted"  gorm:"column:deleted"`
}

type DashEditionVO struct {
	Edition string `json:"edition" gorm:"column:edition"`
	Total   int    `json:"total"   gorm:"column:total"`
}

type DashELevelResp struct {
	Critical int `json:"critical"`
	Major    int `json:"major"`
	Minor    int `json:"minor"`
	Note     int `json:"note"`
}

type DashRLevelResp struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Middle   int `json:"middle"`
	Low      int `json:"low"`
}

type DashRiskstsResp struct {
	Unprocessed int `json:"unprocessed"` // 未处理
	Processed   int `json:"processed"`   // 已处理
	Ignore      int `json:"ignore"`      // 忽略
}
