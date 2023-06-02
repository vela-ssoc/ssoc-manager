package param

type SBOMVulnProject struct {
	Page
	PURL string `json:"purl" query:"purl"`
}
