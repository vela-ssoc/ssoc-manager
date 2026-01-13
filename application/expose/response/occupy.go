package response

type OccupyStat struct {
	FromCode string `json:"from_code"        query:"from_code" gorm:"column:from_code"`
	MinionID int64  `json:"minion_id,string" query:"minion_id" gorm:"column:minion_id"`
	Inet     string `json:"inet"             query:"inet"      gorm:"column:inet"`
	Count    int64  `json:"count"            query:"count"     gorm:"column:count"`
}
