package param

type SIEMServerUpsert struct {
	Name  string `json:"name"  validate:"required,lte=20"`
	URL   string `json:"url"   validate:"required,lte=255"`
	Token string `json:"token" validate:"required,lte=255"`
}
