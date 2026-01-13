package request

import "time"

type OccupyPages struct {
	Pages
	FromCode []string  `json:"from_code"        query:"from_code" validate:"omitempty,lte=100,dive,required"`
	MinionID int64     `json:"minion_id,string" query:"minion_id"`
	OccurAt  time.Time `json:"occur_at"         query:"occur_at"`
}
