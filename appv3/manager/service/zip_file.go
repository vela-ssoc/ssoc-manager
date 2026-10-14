package service

import (
	"archive/zip"
	"bufio"
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/dbzipfs"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-manager/appv3/manager/request"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

type ZipFile struct {
	db  *repository.BaseDB
	log *slog.Logger
}

func NewZipFile(db *repository.BaseDB, log *slog.Logger) *ZipFile {
	return &ZipFile{db: db, log: log}
}

// Unzip 摄入 zip 文件，并按照 zip 在数据库中创建对应的虚拟文件系统。
func (zf *ZipFile) Unzip(ctx context.Context, param request.ZipArchiveCreate, zr *zip.Reader) error {
	if err := zf.precheck(zr); err != nil { // 检测 zip 文件内部信息是否合法
		return err
	}

	// 检查文件是否存在
	coll := zf.db.ZipArchive()
	if _, err := coll.FindByName(ctx, param.Name); err == nil {
		return os.ErrExist
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	now := time.Now()
	archiveID := bson.NewObjectID()
	for _, file := range zr.File {
		if err := zf.saveZipFile(ctx, archiveID, file, now); err != nil {
			return err
		}
	}

	data := &model.ZipArchive{
		ID:          archiveID,
		Name:        param.Name,
		Description: param.Description,
		Method:      param.Method,
		Password:    param.Password,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := coll.InsertOne(ctx, data)
	// TODO 是否要回滚文件内容

	return err
}

// List 获取文件列表。
func (zf *ZipFile) List(ctx context.Context, name string) (*model.ZipArchive, []*model.ZipFile, error) {
	achc, coll := zf.db.ZipArchive(), zf.db.ZipFile()
	opts := options.FindOne().SetProjection(bson.M{"_id": 1})
	archive, err := achc.FindByName(ctx, name, opts)
	if err != nil {
		return nil, nil, err
	}
	filter := bson.M{"archive_id": archive.ID}
	files, err := coll.Find(ctx, filter)

	return archive, files, err
}

// Mkdir 创建目录
func (zf *ZipFile) Mkdir(ctx context.Context, name, fullName string) error {
	coll := zf.db.ZipArchive()
	ach, err := coll.FindByName(ctx, name)
	if err != nil {
		return err
	}

	now := time.Now()
	if err = zf.mkdirAll(ctx, ach.ID, fullName, now); err != nil {
		return err
	}

	// inc 版本号
	increment := bson.M{
		"$inc": bson.M{"version": 1},
		"$set": bson.M{"updated_at": now},
	}
	_, err = coll.UpdateByID(ctx, ach.ID, increment)

	return err
}

// Update 更新文件内容。
func (zf *ZipFile) Update(ctx context.Context, name, fullName string, content io.Reader) error {
	now := time.Now()
	fullName = path.Clean(fullName)
	directory, fileName := path.Split(fullName)
	directory = path.Clean(directory)

	br := bufio.NewReader(content)
	mimeType := zf.detectMIMEType(br, fileName)

	var notexists bool
	archive, file, err := zf.Get(ctx, name, fullName)
	if err != nil {
		if archive == nil {
			return err
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		notexists = file == nil
	}
	if !notexists && file.IsDirectory {
		return os.ErrExist
	}

	coll := zf.db.ZipFile()
	bucket := coll.GridFSBucket()

	// 如果文件不存在，说明首次创建，要先逐级创建父目录，然后上传文件内容。
	if notexists {
		if err = zf.mkdirAll(ctx, archive.ID, directory, now); err != nil {
			return err
		}
	}

	// 上传文件内容
	fileID := bson.NewObjectID()
	size, sum, err := zf.saveContent(ctx, fileID, fullName, br)
	if err != nil {
		return err
	}
	if notexists { // 直接新建文件
		newFile := &model.ZipFile{
			ArchiveID: archive.ID,
			GridID:    fileID,
			Directory: directory,
			FileName:  fileName,
			FullName:  fullName,
			FileSize:  size,
			MIMEType:  mimeType,
			Version:   1,
			Checksum:  sum,
			CreatedAt: now,
			UpdatedAt: now,
		}
		_, err = coll.InsertOne(ctx, newFile)
	} else { // 更新已存在的文件信息
		update := bson.M{
			"$inc": bson.M{"version": 1},
			"$set": bson.M{"grid_id": fileID, "file_size": size, "mime_type": mimeType, "updated_at": now},
		}
		_, err = coll.UpdateByID(ctx, file.ID, update)
	}

	if err != nil {
		_ = bucket.Delete(ctx, fileID) // rollback content
		return err
	} else if file != nil {
		_ = bucket.Delete(ctx, file.GridID) // remove old content
	}

	// 归档文件信息版本号 +1
	if notexists || file.Checksum.SHA512 != sum.SHA512 {
		increment := bson.M{
			"$inc": bson.M{"version": 1},
			"$set": bson.M{"updated_at": now},
		}
		achc := zf.db.ZipArchive()
		_, err = achc.UpdateByID(ctx, archive.ID, increment)
	}

	return err
}

// Delete 删除文件。
func (zf *ZipFile) Delete(ctx context.Context, name, fullName string) error {
	archive, _, err := zf.Get(ctx, name, fullName)
	if err != nil {
		return err
	}

	// 文件类型删除文件和内容即可。
	// 目录类型要删除当前目录和该目录下所有子目录和子目录下的文件。
	zfs := dbzipfs.NewZip(ctx, zf.db, archive.ID)
	err = zf.deleteAll(ctx, zfs, fullName)

	return err
}

func (zf *ZipFile) Get(ctx context.Context, name, fullName string) (*model.ZipArchive, *model.ZipFile, error) {
	achc, coll := zf.db.ZipArchive(), zf.db.ZipFile()
	opts := options.FindOne().SetProjection(bson.M{"_id": 1})
	archive, err := achc.FindByName(ctx, name, opts)
	if err != nil {
		return nil, nil, err
	}

	fullName = path.Clean(fullName)
	filter := bson.M{"archive_id": archive.ID, "full_name": fullName}
	file, err := coll.FindOne(ctx, filter)

	return archive, file, err
}

func (zf *ZipFile) Open(ctx context.Context, name, fullName string) (*model.ZipArchive, *model.ZipFile, *mongo.GridFSDownloadStream, error) {
	archive, file, err := zf.Get(ctx, name, fullName)
	if err != nil {
		return archive, file, nil, err
	}

	bucket := zf.db.ZipFile().GridFSBucket()
	stm, err := bucket.OpenDownloadStream(ctx, file.GridID)

	return archive, file, stm, err
}

func (zf *ZipFile) deleteAll(ctx context.Context, zfs *dbzipfs.Zip, root string) error {
	coll := zf.db.ZipFile()
	bucket := coll.GridFSBucket()

	return fs.WalkDir(zfs, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		file, yes := info.(*model.ZipFile)
		if !yes {
			return nil
		}
		if _, err = coll.DeleteByID(ctx, file.ID); err != nil {
			return err
		}
		if !file.GridID.IsZero() {
			err = bucket.Delete(ctx, file.GridID)
		}

		return err
	})
}

func (zf *ZipFile) mkdirAll(ctx context.Context, archiveID bson.ObjectID, dir string, createdAt time.Time) error {
	dir = path.Clean(dir)
	coll := zf.db.ZipFile()
	for {
		if dir == "" || dir == "." {
			return nil
		}

		filter := bson.M{"archive_id": archiveID, "full_name": dir}
		last, err := coll.FindOne(ctx, filter)
		if last != nil {
			// 创建目录前，先检查目录是否存在。
			if !last.IsDirectory {
				return os.ErrExist // 文件已存在
			}
			return nil
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}

		directory, fileName := path.Split(dir)
		directory = path.Clean(directory)
		data := &model.ZipFile{
			ArchiveID:   archiveID,
			Directory:   directory,
			FileName:    fileName,
			FullName:    dir,
			IsDirectory: true,
			Version:     1,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}
		if _, err = coll.InsertOne(ctx, data); err != nil {
			return err
		}

		dir = directory
	}
}

//goland:noinspection GoUnhandledErrorResult
func (zf *ZipFile) saveContent(ctx context.Context, fileID bson.ObjectID, name string, r io.Reader) (int64, model.Checksum, error) {
	h := model.NewHashWriter()
	tr := io.TeeReader(r, h)

	bucket := zf.db.ZipFile().GridFSBucket()
	stm, err := bucket.OpenUploadStreamWithID(ctx, fileID, name)
	if err != nil {
		return 0, model.Checksum{}, nil
	}
	defer stm.Close()

	size, err := io.Copy(stm, tr)
	if err != nil {
		return size, model.Checksum{}, nil
	}
	sum := h.Sum()

	return size, sum, nil
}

func (zf *ZipFile) saveZipFile(ctx context.Context, archiveID bson.ObjectID, file *zip.File, now time.Time) error {
	fullName := file.Name
	if file.NonUTF8 || !utf8.ValidString(fullName) {
		str, ok := zf.transcode(fullName)
		if !ok {
			return encoding.ErrInvalidUTF8
		}
		fullName = str
	}
	fullName = path.Clean(fullName)
	info := file.FileInfo()
	if info.IsDir() {
		return zf.mkdirAll(ctx, archiveID, fullName, now)
	}

	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	br := bufio.NewReader(rc)
	mimeType := zf.detectMIMEType(br, fullName)

	directory, fileName := path.Split(fullName)
	directory = path.Clean(directory)
	if err = zf.mkdirAll(ctx, archiveID, directory, now); err != nil {
		return err
	}

	fileID := bson.NewObjectID()
	size, sum, err := zf.saveContent(ctx, fileID, fullName, br)
	if err != nil {
		return err
	}

	data := &model.ZipFile{
		ArchiveID: archiveID,
		GridID:    fileID,
		Directory: directory,
		FileName:  fileName,
		FullName:  fullName,
		FileSize:  size,
		MIMEType:  mimeType,
		Version:   1,
		Checksum:  sum,
		CreatedAt: now,
		UpdatedAt: now,
	}
	coll := zf.db.ZipFile()
	if _, err = coll.InsertOne(ctx, data); err != nil {
		bucket := coll.GridFSBucket()
		_ = bucket.Delete(ctx, fileID) // rollback content
	}

	return err
}

func (zf *ZipFile) detectMIMEType(br *bufio.Reader, fileName string) string {
	peek, err := br.Peek(512)
	if err != nil && err != io.EOF {
		return "application/octet-stream"
	}

	str := http.DetectContentType(peek)
	if str != "application/octet-stream" { // fallback
		return str
	}

	ext := strings.ToLower(path.Ext(fileName))
	if typ := mime.TypeByExtension(ext); typ != "" {
		return typ
	}

	return str
}

func (zf *ZipFile) precheck(zr *zip.Reader) error {
	return nil
}

func (zf *ZipFile) transcode(str string) (string, bool) {
	coders := []encoding.Encoding{
		simplifiedchinese.GBK,
		traditionalchinese.Big5,
		japanese.EUCJP,
		korean.EUCKR,
	}
	for _, coder := range coders {
		if val, err := coder.NewDecoder().String(str); err == nil {
			return val, true
		}
	}

	return "", false
}
