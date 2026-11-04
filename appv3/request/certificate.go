package request

type CertificateCreate struct {
	PublicKey  string `json:"public_key"  validate:"required"`
	PrivateKey string `json:"private_key" validate:"required"`
}

type CertificateUpdate struct {
	ObjectID
	PublicKey  string `json:"public_key"  validate:"required"`
	PrivateKey string `json:"private_key" validate:"required"`
}
