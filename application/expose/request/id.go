package request

import "go.mongodb.org/mongo-driver/v2/bson"

type ObjectID struct {
	ID string `json:"id" query:"id" validate:"required,hex"`
}

func (i ObjectID) Get() bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(i.ID)
	return id
}
