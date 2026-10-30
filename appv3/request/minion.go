package request

type MinionPage struct {
	PageSizeKeyword
	Department string `json:"department" query:"department"`
}
