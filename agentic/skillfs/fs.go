package skillfs

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cloudwego/eino/adk/filesystem"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Node 是虚拟文件系统中的一个节点，等价于 zip 里的一条目录项
// （对应 zip_file 集合的 file_name/full_name/is_directory/grid_id 语义）。
//
// 路径统一使用 POSIX 风格、以 / 开头的绝对路径，例如：
//
//	/skills/log_analyzer/SKILL.md
//
// 文本文件正文内联在 Content 中（便于 grep 直接在库里做正则匹配），
// 二进制或超大文件落到 GridFS，GridID 指向 blob。
type Node struct {
	ID        bson.ObjectID `bson:"_id,omitempty"        json:"id"`
	Path      string        `bson:"path"                 json:"path"`                // 绝对路径
	Parent    string        `bson:"parent"               json:"parent"`              // 父目录绝对路径，根目录的子节点为 "/"
	Name      string        `bson:"name"                 json:"name"`                // 末级名称
	IsDir     bool          `bson:"is_dir"               json:"is_dir"`              // 是否目录节点
	Size      int64         `bson:"size"                 json:"size"`                // 字节数
	MIME      string        `bson:"mime_type,omitempty"  json:"mime_type,omitempty"` // MIME，例如 text/markdown
	Content   string        `bson:"content,omitempty"    json:"content,omitempty"`   // 文本正文
	GridID    bson.ObjectID `bson:"grid_id,omitempty"    json:"grid_id,omitempty"`   // 二进制内容
	CreatedAt time.Time     `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty" json:"updated_at"`
}

// FS 是基于 MongoDB 的虚拟文件系统，实现 eino 的 filesystem.Backend。
//
// 它同时可以喂给 skill.NewBackendFromFilesystem 加载技能，也可以作为 agent
// 读写脚本的工具后端，两边看到的是同一棵树。
type FS struct {
	coll   *mongo.Collection
	bucket *mongo.GridFSBucket
}

var _ filesystem.Backend = (*FS)(nil)

func newFS(db *mongo.Database, collName, bucketName string) *FS {
	return &FS{
		coll:   db.Collection(collName),
		bucket: db.GridFSBucket(options.GridFSBucket().SetName(bucketName)),
	}
}

// EnsureIndex 创建 path 唯一索引与目录列表索引。
// 重复执行无害，建议在启动时调用一次。
func (f *FS) EnsureIndex(ctx context.Context, opts ...options.Lister[options.CreateIndexesOptions]) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "path", Value: 1}},
			Options: options.Index().SetName("uniq_path").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "parent", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetName("parent_name"),
		},
	}
	_, err := f.coll.Indexes().CreateMany(ctx, indexes, opts...)

	return err
}

// Collection 暴露底层集合，便于管理端做分页浏览等扩展查询。
func (f *FS) Collection() *mongo.Collection { return f.coll }

// LsInfo 列出目录下一级节点。
func (f *FS) LsInfo(ctx context.Context, req *filesystem.LsInfoRequest) ([]filesystem.FileInfo, error) {
	dir := normalize("")
	if req != nil && req.Path != "" {
		dir = normalize(req.Path)
	}

	opt := options.Find().SetSort(bson.D{
		{Key: "is_dir", Value: -1},
		{Key: "name", Value: 1},
	})
	nodes, err := f.find(ctx, bson.D{{Key: "parent", Value: dir}}, opt)
	if err != nil {
		return nil, fmt.Errorf("skillfs: ls %s: %w", dir, err)
	}

	ret := make([]filesystem.FileInfo, 0, len(nodes))
	for _, node := range nodes {
		ret = append(ret, node.fileInfo())
	}

	return ret, nil
}

// Read 读取文件内容，支持按行 offset/limit 分片。
func (f *FS) Read(ctx context.Context, req *filesystem.ReadRequest) (*filesystem.FileContent, error) {
	if req == nil {
		return nil, errors.New("skillfs: request is required")
	}

	node, err := f.stat(ctx, normalize(req.FilePath))
	if err != nil {
		return nil, err
	}
	if node.IsDir {
		return nil, fmt.Errorf("skillfs: %s is a directory", node.Path)
	}

	data, err := f.load(ctx, node)
	if err != nil {
		return nil, err
	}

	return &filesystem.FileContent{Content: sliceLines(string(data), req.Offset, req.Limit)}, nil
}

// GrepRaw 在指定目录下做正则搜索。只扫描内联文本，二进制文件不参与。
func (f *FS) GrepRaw(ctx context.Context, req *filesystem.GrepRequest) ([]filesystem.GrepMatch, error) {
	if req == nil || req.Pattern == "" {
		return nil, errors.New("skillfs: grep pattern is required")
	}

	expr := req.Pattern
	if req.CaseInsensitive {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("skillfs: invalid grep pattern: %w", err)
	}

	scope := normalize("")
	if req.Path != "" {
		scope = normalize(req.Path)
	}

	filter := bson.D{
		{Key: "is_dir", Value: false},
		{Key: "grid_id", Value: bson.D{{Key: "$exists", Value: false}}},
	}
	filter = append(filter, underScope(scope)...)

	nodes, err := f.find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("skillfs: grep %s: %w", scope, err)
	}

	fileType := strings.TrimPrefix(strings.ToLower(req.FileType), ".")
	matches := make([]filesystem.GrepMatch, 0, 16)
	for _, node := range nodes {
		rel := relative(node.Path, scope)
		if req.Glob != "" {
			ok, matchErr := doublestar.Match(req.Glob, rel)
			if matchErr != nil {
				return nil, fmt.Errorf("skillfs: invalid grep glob %q: %w", req.Glob, matchErr)
			}
			if !ok {
				continue
			}
		}
		if fileType != "" && !strings.EqualFold(path.Ext(node.Path), "."+fileType) {
			continue
		}

		if req.EnableMultiline {
			matches = appendMultiline(matches, node.Path, node.Content, re)
		} else {
			matches = appendLineMatches(matches, node.Path, node.Content, re)
		}
	}

	return matches, nil
}

// GlobInfo 按 glob 匹配路径，支持 **。Path 为基准目录，Pattern 相对它匹配。
func (f *FS) GlobInfo(ctx context.Context, req *filesystem.GlobInfoRequest) ([]filesystem.FileInfo, error) {
	if req == nil || req.Pattern == "" {
		return nil, errors.New("skillfs: glob pattern is required")
	}

	scope := normalize("")
	if req.Path != "" {
		scope = normalize(req.Path)
	}

	nodes, err := f.find(ctx, underScope(scope))
	if err != nil {
		return nil, fmt.Errorf("skillfs: glob %s: %w", scope, err)
	}

	ret := make([]filesystem.FileInfo, 0, 8)
	for _, node := range nodes {
		rel := relative(node.Path, scope)
		if rel == "" {
			continue
		}
		ok, matchErr := doublestar.Match(req.Pattern, rel)
		if matchErr != nil {
			return nil, fmt.Errorf("skillfs: invalid glob %q: %w", req.Pattern, matchErr)
		}
		if ok {
			ret = append(ret, node.fileInfo())
		}
	}

	return ret, nil
}

// Write 创建或覆盖文本文件，父目录不存在时自动补齐。
func (f *FS) Write(ctx context.Context, req *filesystem.WriteRequest) error {
	if req == nil {
		return errors.New("skillfs: request is required")
	}

	return f.put(ctx, normalize(req.FilePath), []byte(req.Content), time.Now())
}

// Edit 对文本文件做字面替换。
func (f *FS) Edit(ctx context.Context, req *filesystem.EditRequest) error {
	if req == nil || req.OldString == "" {
		return errors.New("skillfs: old string is required")
	}
	if req.OldString == req.NewString {
		return errors.New("skillfs: new string must differ from old string")
	}

	filePath := normalize(req.FilePath)
	node, err := f.stat(ctx, filePath)
	if err != nil {
		return err
	}
	if node.IsDir {
		return fmt.Errorf("skillfs: %s is a directory", filePath)
	}

	data, err := f.load(ctx, node)
	if err != nil {
		return err
	}
	content := string(data)

	switch count := strings.Count(content, req.OldString); {
	case count == 0:
		return fmt.Errorf("skillfs: old string not found in %s", filePath)
	case count > 1 && !req.ReplaceAll:
		return fmt.Errorf("skillfs: old string appears %d times in %s, set ReplaceAll", count, filePath)
	}

	var replaced string
	if req.ReplaceAll {
		replaced = strings.ReplaceAll(content, req.OldString, req.NewString)
	} else {
		replaced = strings.Replace(content, req.OldString, req.NewString, 1)
	}

	return f.put(ctx, filePath, []byte(replaced), time.Now())
}

// Mkdir 创建目录（含上级），已存在则忽略。
func (f *FS) Mkdir(ctx context.Context, dir string) error {
	return f.ensureDirs(ctx, normalize(dir))
}

// Remove 删除文件或目录（目录连同子孙一起删，blob 一并清理）。
func (f *FS) Remove(ctx context.Context, target string) error {
	target = normalize(target)
	if target == "/" {
		return errors.New("skillfs: cannot remove root")
	}

	filter := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "path", Value: target}},
		bson.D{{Key: "path", Value: bson.D{{Key: "$regex", Value: "^" + regexp.QuoteMeta(target) + "/"}}}},
	}}}
	nodes, err := f.find(ctx, filter)
	if err != nil {
		return fmt.Errorf("skillfs: remove %s: %w", target, err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("%w: %s", fs.ErrNotExist, target)
	}

	for _, node := range nodes {
		if node.GridID.IsZero() {
			continue
		}
		if delErr := f.bucket.Delete(ctx, node.GridID); delErr != nil {
			return fmt.Errorf("skillfs: delete blob %s: %w", node.Path, delErr)
		}
	}

	if _, err = f.coll.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("skillfs: remove %s: %w", target, err)
	}

	return nil
}

// ImportZip 把 zip 的目录层级原样铺到 dir 下，返回导入的文件数。
// 与 zip 的行为一致：目录即目录，文件即文件，多级嵌套全部保留。
func (f *FS) ImportZip(ctx context.Context, dir string, data []byte) (int, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("skillfs: open zip: %w", err)
	}

	base := normalize(dir)
	count := 0
	for _, entry := range zr.File {
		name := strings.TrimPrefix(strings.ReplaceAll(entry.Name, "\\", "/"), "./")
		if name == "" || strings.HasPrefix(name, "__MACOSX/") || strings.HasSuffix(name, "/.DS_Store") {
			continue
		}

		target := normalize(path.Join(base, name))
		if entry.FileInfo().IsDir() {
			if err = f.ensureDirs(ctx, target); err != nil {
				return count, err
			}
			continue
		}

		rc, openErr := entry.Open()
		if openErr != nil {
			return count, fmt.Errorf("skillfs: open zip entry %s: %w", entry.Name, openErr)
		}
		content, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			return count, fmt.Errorf("skillfs: read zip entry %s: %w", entry.Name, readErr)
		}

		if err = f.put(ctx, target, content, entry.Modified); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

func (f *FS) find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Node, error) {
	cur, err := f.coll.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()

	nodes := make([]*Node, 0, 16)
	if err = cur.All(ctx, &nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (f *FS) stat(ctx context.Context, target string) (*Node, error) {
	node := new(Node)
	err := f.coll.FindOne(ctx, bson.D{{Key: "path", Value: target}}).Decode(node)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: %s", fs.ErrNotExist, target)
		}
		return nil, fmt.Errorf("skillfs: stat %s: %w", target, err)
	}

	return node, nil
}

func (f *FS) load(ctx context.Context, node *Node) ([]byte, error) {
	if node.GridID.IsZero() {
		return []byte(node.Content), nil
	}

	stm, err := f.bucket.OpenDownloadStream(ctx, node.GridID)
	if err != nil {
		return nil, fmt.Errorf("skillfs: open blob %s: %w", node.Path, err)
	}
	defer func() { _ = stm.Close() }()

	data, err := io.ReadAll(stm)
	if err != nil {
		return nil, fmt.Errorf("skillfs: read blob %s: %w", node.Path, err)
	}

	return data, nil
}

// put 写入单个文件节点：UTF-8 文本内联，二进制走 GridFS。
func (f *FS) put(ctx context.Context, target string, data []byte, mod time.Time) error {
	if target == "/" {
		return errors.New("skillfs: cannot write root")
	}
	if err := f.ensureDirs(ctx, path.Dir(target)); err != nil {
		return err
	}

	if mod.IsZero() {
		mod = time.Now()
	}

	set := bson.D{
		{Key: "path", Value: target},
		{Key: "parent", Value: path.Dir(target)},
		{Key: "name", Value: path.Base(target)},
		{Key: "is_dir", Value: false},
		{Key: "size", Value: int64(len(data))},
		{Key: "mime_type", Value: mimeOf(target)},
		{Key: "updated_at", Value: mod},
	}
	unset := bson.D{}
	if utf8.Valid(data) {
		set = append(set, bson.E{Key: "content", Value: string(data)})
		unset = append(unset, bson.E{Key: "grid_id", Value: ""})
	} else {
		id, err := f.bucket.UploadFromStream(ctx, path.Base(target), bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("skillfs: upload blob %s: %w", target, err)
		}
		set = append(set, bson.E{Key: "grid_id", Value: id})
		unset = append(unset, bson.E{Key: "content", Value: ""})
	}

	update := bson.D{
		{Key: "$set", Value: set},
		{Key: "$setOnInsert", Value: bson.D{{Key: "created_at", Value: time.Now()}}},
	}
	if len(unset) > 0 {
		update = append(update, bson.E{Key: "$unset", Value: unset})
	}

	filter := bson.D{{Key: "path", Value: target}}
	opt := options.UpdateOne().SetUpsert(true)
	if _, err := f.coll.UpdateOne(ctx, filter, update, opt); err != nil {
		return fmt.Errorf("skillfs: write %s: %w", target, err)
	}

	return nil
}

// ensureDirs 逐级补齐目录节点，已存在的目录不会被覆盖。
func (f *FS) ensureDirs(ctx context.Context, dir string) error {
	dir = normalize(dir)
	if dir == "/" {
		return nil
	}

	now := time.Now()
	parts := strings.Split(strings.TrimPrefix(dir, "/"), "/")
	models := make([]mongo.WriteModel, 0, len(parts))

	parent, current := "/", ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		current += "/" + part
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "path", Value: current}}).
			SetUpdate(bson.D{{Key: "$setOnInsert", Value: bson.D{
				{Key: "path", Value: current},
				{Key: "parent", Value: parent},
				{Key: "name", Value: part},
				{Key: "is_dir", Value: true},
				{Key: "created_at", Value: now},
				{Key: "updated_at", Value: now},
			}}}).
			SetUpsert(true))
		parent = current
	}

	if len(models) == 0 {
		return nil
	}
	if _, err := f.coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(true)); err != nil {
		return fmt.Errorf("skillfs: mkdir %s: %w", dir, err)
	}

	return nil
}

// normalize 统一成 POSIX 绝对路径，同时兼容调用方传入 Windows 分隔符。
func normalize(p string) string {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	if p == "" {
		return "/"
	}

	return path.Clean("/" + strings.TrimPrefix(p, "/"))
}

// underScope 生成“基准目录之下（不含自身）”的过滤条件。
func underScope(scope string) bson.D {
	if scope == "/" {
		return bson.D{}
	}

	prefix := bson.D{{Key: "$regex", Value: "^" + regexp.QuoteMeta(scope) + "/"}}
	return bson.D{{Key: "path", Value: prefix}}
}

// relative 返回 node 相对 scope 的路径。
func relative(nodePath, scope string) string {
	if scope == "/" {
		return strings.TrimPrefix(nodePath, "/")
	}

	return strings.TrimPrefix(strings.TrimPrefix(nodePath, scope), "/")
}

func (n *Node) fileInfo() filesystem.FileInfo {
	modified := ""
	if !n.UpdatedAt.IsZero() {
		modified = n.UpdatedAt.UTC().Format(time.RFC3339)
	}

	return filesystem.FileInfo{
		Path:       n.Path,
		IsDir:      n.IsDir,
		Size:       n.Size,
		ModifiedAt: modified,
	}
}

func appendLineMatches(dst []filesystem.GrepMatch, filePath, content string, re *regexp.Regexp) []filesystem.GrepMatch {
	for idx, line := range strings.Split(content, "\n") {
		if !re.MatchString(line) {
			continue
		}
		dst = append(dst, filesystem.GrepMatch{Content: line, Path: filePath, Line: idx + 1})
	}

	return dst
}

func appendMultiline(dst []filesystem.GrepMatch, filePath, content string, re *regexp.Regexp) []filesystem.GrepMatch {
	for _, loc := range re.FindAllStringIndex(content, -1) {
		line := 1 + strings.Count(content[:loc[0]], "\n")
		dst = append(dst, filesystem.GrepMatch{
			Content: content[loc[0]:loc[1]],
			Path:    filePath,
			Line:    line,
		})
	}

	return dst
}

// sliceLines 按 1-based 行号切片，offset < 1 视为 1，limit <= 0 表示读到底。
func sliceLines(content string, offset, limit int) string {
	if offset <= 1 && limit <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	if offset < 1 {
		offset = 1
	}
	if offset > len(lines) {
		return ""
	}

	start := offset - 1
	end := len(lines)
	if limit > 0 && start+limit < end {
		end = start + limit
	}

	return strings.Join(lines[start:end], "\n")
}

func mimeOf(p string) string {
	if t := mime.TypeByExtension(path.Ext(p)); t != "" {
		return t
	}

	return "application/octet-stream"
}
