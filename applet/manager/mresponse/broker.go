package mresponse

type BrokerOnline struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
	Goos string `json:"goos"`
	Arch string `json:"arch"`
}
