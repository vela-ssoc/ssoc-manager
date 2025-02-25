package mrequest

import "mime/multipart"

type GridUpload struct {
	Name string                `form:"name" validate:"required,lte=255"`
	File *multipart.FileHeader `form:"file" validate:"required"`
}
