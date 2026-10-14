package request

import "mime/multipart"

type ZipFileUpdate struct {
	ZipFileName
	File *multipart.FileHeader `json:"file" form:"file" validate:"required"`
}

type ZipFileName struct {
	Name     string `json:"name"      query:"name"      form:"name"      param:"name"      validate:"required"`
	FullName string `json:"full_name" query:"full_name" form:"full_name" param:"full_name" validate:"required"`
}

type ZipFileUnzip struct {
	ZipArchiveCreate
	File *multipart.FileHeader `json:"file" form:"file" validate:"required"`
}
