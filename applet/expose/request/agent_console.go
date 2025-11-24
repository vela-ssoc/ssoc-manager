package request

type AgentConsoleRead struct {
	ID int64 `json:"id" form:"id" query:"id" validate:"required"`
	N  int   `json:"n"  form:"n"  query:"n"`
}

type AgentConsoleStat struct {
	Size    int64 `json:"size"`
	Maxsize int64 `json:"maxsize"`
}
