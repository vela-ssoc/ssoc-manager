package request

import "go.mongodb.org/mongo-driver/v2/bson"

type ObjectID struct {
	ID string `json:"id" query:"id" form:"id" param:"id" validate:"required,mongodb"`
}

func (o ObjectID) OID() bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(o.ID)
	return id
}

type Name struct {
	Name string `json:"name" query:"name" form:"name" param:"name" validate:"required"`
}

type NamePath struct {
	Name string `json:"name" query:"name" form:"name" param:"name" validate:"required"`
	Path string `json:"path" query:"path" form:"path" param:"path" validate:"required"`
}
