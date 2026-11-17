package agtool

import (
	"cmp"
	"context"
	"slices"
)

type CollectionsInput struct{}

type CollectionInfo struct {
	Collection  string `json:"collection"`
	Description string `json:"description"`
}

type CollectionsOutput struct {
	Collections []CollectionInfo `json:"collections"`
}

func (mgo *mongoInstance) collections(context.Context, emptyStruct) (*CollectionsOutput, error) {
	items := make([]CollectionInfo, 0, len(mgo.colls))
	for _, mod := range mgo.colls {
		items = append(items, mod)
	}
	slices.SortFunc(items, func(a, b CollectionInfo) int {
		return cmp.Compare(a.Collection, b.Collection)
	})

	return &CollectionsOutput{Collections: items}, nil
}
