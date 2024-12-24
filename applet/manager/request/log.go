package request

type LogChange struct {
	Log string `json:"log" validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
	ORM string `json:"orm" validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
}
