package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/dbzipfs"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ZipArchive struct {
	db  polydb.Sourcer
	log *slog.Logger
}

func NewZipArchive(db polydb.Sourcer, log *slog.Logger) *ZipArchive {
	return &ZipArchive{db: db, log: log}
}

func (za *ZipArchive) Page(ctx context.Context, param request.PageSizeKeyword) (*repository.Pages[model.ZipArchive], error) {
	filter := param.RegexTo(nil, "name", "description")
	coll := za.db.MDB().ZipArchive()
	pn, sn := param.PSN()

	return coll.Page(ctx, filter, pn, sn)
}

func (za *ZipArchive) Get(ctx context.Context, name string) (*model.ZipArchive, error) {
	coll := za.db.MDB().ZipArchive()
	return coll.FindByName(ctx, name)
}

func (za *ZipArchive) Open(ctx context.Context, name string) (*model.ZipArchive, *dbzipfs.Zip, error) {
	dat, err := za.Get(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	mdb := za.db.MDB()
	zfs := dbzipfs.NewZip(ctx, mdb, dat.ID)

	return dat, zfs, nil
}

// Create 创建一个空的虚拟 zip 归档。
func (za *ZipArchive) Create(ctx context.Context, param request.ZipArchiveCreate) error {
	now := time.Now()
	data := &model.ZipArchive{
		Name:        param.Name,
		Description: param.Description,
		Method:      param.Method,
		Password:    param.Password,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	coll := za.db.MDB().ZipArchive()
	_, err := coll.InsertOne(ctx, data)

	return err
}

// Update 更新虚拟 zip 归档文件的基本信息。
func (za *ZipArchive) Update(ctx context.Context, param request.ZipArchiveUpdate) error {
	now := time.Now()
	update := bson.M{
		"$inc": bson.M{"version": 1},
		"$set": bson.M{
			"description": param.Description,
			"method":      param.Method,
			"password":    param.Password,
			"updated_at":  now,
		},
	}

	coll := za.db.MDB().ZipArchive()
	_, err := coll.UpdateOne(ctx, bson.M{"name": param.Name}, update)

	return err
}
