package request

type ZipArchiveCreate struct {
	Name        string `json:"name"        form:"name"        validate:"required,lte=100"`
	Description string `json:"description" form:"description" validate:"lte=1000"`
	Method      int    `json:"method"      form:"method"      validate:"gte=0,lte=4"` // 0-不加密 1-ZipCrypto 2-AES-128 3-AES-192 4-AES-256
	Password    string `json:"password"    form:"password"    validate:"lte=100"`
}

type ZipArchiveUpdate struct {
	ObjectID
	ZipArchiveCreate
}
