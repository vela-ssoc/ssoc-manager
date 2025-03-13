package mresponse

type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type NameValues []*NameValue
