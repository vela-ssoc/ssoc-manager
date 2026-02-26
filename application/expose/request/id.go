package request

import "go.mongodb.org/mongo-driver/v2/bson"

type HexID struct {
	ID string `json:"id" form:"id" query:"id" validate:"required,mongodb"`
}

func (h HexID) ObjectID() (bson.ObjectID, error) {
	return bson.ObjectIDFromHex(h.ID)
}

// MustID 虽然叫 Must 但是错误不会 panic。
func (h HexID) MustID() bson.ObjectID {
	id, _ := h.ObjectID()
	return id
}

type HexIDs struct {
	ID []string `json:"id" form:"id" query:"id" validate:"unique,dive,required,mongodb"`
}

func (h HexIDs) MustIDs() []bson.ObjectID {
	var ids []bson.ObjectID
	for _, s := range h.ID {
		if id, err := bson.ObjectIDFromHex(s); err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}

func (h HexIDs) ObjectIDs() ([]bson.ObjectID, error) {
	var ids []bson.ObjectID
	for _, s := range h.ID {
		id, err := bson.ObjectIDFromHex(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
