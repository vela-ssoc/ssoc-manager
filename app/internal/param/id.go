package param

import (
	"encoding/json"
	"strconv"
)

type IntID struct {
	ID int64 `json:"id,string" form:"id" query:"id" validate:"required,gt=0"`
}

type OptionalIDs struct {
	ID Int64s `json:"id" form:"id" query:"id"`
}

type Int64s []int64

func (is Int64s) MarshalJSON() (text []byte, err error) {
	size := len(is)
	str := make([]string, 0, size)
	for _, i := range is {
		s := strconv.FormatInt(i, 10)
		str = append(str, s)
	}

	return json.Marshal(str)
}

func (is *Int64s) UnmarshalJSON(raw []byte) error {
	var str []string
	if err := json.Unmarshal(raw, &str); err != nil {
		return err
	}

	dats := make([]int64, 0, len(str))
	for _, st := range str {
		num, err := strconv.ParseInt(st, 10, 64)
		if err != nil {
			return err
		}
		dats = append(dats, num)
	}
	*is = dats

	return nil
}

type Data struct {
	Data any `json:"data,omitempty"`
}

func WarpData(dats any) *Data {
	return &Data{Data: dats}
}
