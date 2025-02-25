package mrequest

type LogChange struct {
	Log string `json:"log" validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
	ORM string `json:"orm" validate:"omitempty,oneof=INFO WARN ERROR"`
}
