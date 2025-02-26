package mrequest

type BlinkAlert struct {
	UserIDs  []string `json:"user_ids"`
	GroupIDs []string `json:"group_ids"`
	Detail   string   `json:"detail"`
	Title    string   `json:"title"`
	Text     string   `json:"text"`
}
