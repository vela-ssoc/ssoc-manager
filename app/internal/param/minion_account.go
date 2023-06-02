package param

type MinionAccountPage struct {
	Page
	Name     string `json:"name"             query:"name"`
	MinionID int64  `json:"minion_id,string" query:"minion_id"`
}
