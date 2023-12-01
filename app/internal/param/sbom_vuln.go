package param

type SBOMVulnProject struct {
	Page
	PURL string `json:"purl" query:"purl"`
}

type ReportPurl struct {
	Purl []string `json:"purl" validate:"lte=1000,dive,required,lte=1000"`
}
