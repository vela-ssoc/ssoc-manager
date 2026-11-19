package fskill

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cloudwego/eino/adk/filesystem"
)

// Dir 是一台只读的本地目录文件系统，Dir 的值为根目录。
// Write/Edit 一律拒绝，写操作只走管理端。
type Dir string

const (
	maxReadSize    = 4 << 20 // 单文件最多读 4MB，防止把上下文冲爆
	maxGrepFiles   = 2000    // 单次 grep 最多扫 2000 个文件
	maxGrepMatches = 500     // 单次 grep 最多返回 500 条
)

// root 返回绝对化后的根目录。
func (d Dir) root() string {
	root := strings.TrimSpace(string(d))
	if root == "" {
		root = "."
	}
	if abs, err := filepath.Abs(root); err == nil {
		return abs
	}

	return filepath.Clean(root)
}

// resolve 把入参解析成宿主绝对路径，并保证不会逃出根目录。
// 相对路径按根目录展开；绝对路径必须落在根目录内。
func (d Dir) resolve(name string) (string, error) {
	root := d.root()
	p := strings.TrimSpace(name)
	if p == "" {
		return root, nil
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, filepath.FromSlash(p))
	}
	p = filepath.Clean(p)

	rel, err := filepath.Rel(root, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}

	return p, nil
}

// rel 返回相对根目录、斜杠分隔的路径（对外统一 "/"）。
func (d Dir) rel(p string) string {
	r, err := filepath.Rel(d.root(), p)
	if err != nil {
		return filepath.ToSlash(p)
	}

	return filepath.ToSlash(r)
}

func toFileInfo(rel string, info fs.FileInfo) filesystem.FileInfo {
	return filesystem.FileInfo{
		Path:       rel,
		IsDir:      info.IsDir(),
		Size:       info.Size(),
		ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
	}
}

// LsInfo 列出目录下一级条目，Path 相对根目录。
func (d Dir) LsInfo(_ context.Context, req *filesystem.LsInfoRequest) ([]filesystem.FileInfo, error) {
	dir := d.root()
	if req != nil && req.Path != "" {
		p, err := d.resolve(req.Path)
		if err != nil {
			return nil, err
		}
		dir = p
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil // 目录不存在按空目录处理，与官方 local backend 一致
		}
		return nil, fmt.Errorf("localfs: ls %s: %w", dir, err)
	}

	ret := make([]filesystem.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		ret = append(ret, toFileInfo(d.rel(filepath.Join(dir, entry.Name())), info))
	}
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].IsDir != ret[j].IsDir {
			return ret[i].IsDir
		}
		return ret[i].Path < ret[j].Path
	})

	return ret, nil
}

// Read 读取文本文件，支持按行 offset/limit 分片。
func (d Dir) Read(_ context.Context, req *filesystem.ReadRequest) (*filesystem.FileContent, error) {
	if req == nil {
		return nil, errors.New("localfs: request is required")
	}

	p, err := d.resolve(req.FilePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(p)
	if err != nil {
		return nil, &fs.PathError{Op: "read", Path: req.FilePath, Err: err}
	}
	if info.IsDir() {
		return nil, &fs.PathError{Op: "read", Path: req.FilePath, Err: errors.New("is a directory")}
	}
	if info.Size() > maxReadSize {
		return nil, fmt.Errorf("localfs: %s is too large (%d bytes, limit %d)", req.FilePath, info.Size(), maxReadSize)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("localfs: read %s: %w", req.FilePath, err)
	}

	return &filesystem.FileContent{Content: sliceLines(string(data), req.Offset, req.Limit)}, nil
}

// GlobInfo 按 glob 匹配路径，支持 **。
//
// 注意：返回的 Path 是相对 req.Path 的路径（斜杠分隔），
// skill.NewBackendFromFilesystem 依赖这个约定去拼绝对路径。
func (d Dir) GlobInfo(_ context.Context, req *filesystem.GlobInfoRequest) ([]filesystem.FileInfo, error) {
	if req == nil || req.Pattern == "" {
		return nil, errors.New("localfs: glob pattern is required")
	}

	base := d.root()
	if req.Path != "" {
		p, err := d.resolve(req.Path)
		if err != nil {
			return nil, err
		}
		base = p
	}

	var ret []filesystem.FileInfo
	err := filepath.WalkDir(base, func(p string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == base {
			return nil
		}

		rel, relErr := filepath.Rel(base, p)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)

		ok, matchErr := doublestar.Match(req.Pattern, rel)
		if matchErr != nil {
			return fmt.Errorf("localfs: invalid glob %q: %w", req.Pattern, matchErr)
		}
		if !ok {
			return nil
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil
		}
		ret = append(ret, toFileInfo(rel, info))

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("localfs: glob %s: %w", req.Pattern, err)
	}

	sort.Slice(ret, func(i, j int) bool { return ret[i].Path < ret[j].Path })

	return ret, nil
}

// GrepRaw 在指定目录下做正则搜索，跳过二进制与超大文件。
func (d Dir) GrepRaw(ctx context.Context, req *filesystem.GrepRequest) ([]filesystem.GrepMatch, error) {
	if req == nil || req.Pattern == "" {
		return nil, errors.New("localfs: grep pattern is required")
	}

	expr := req.Pattern
	if req.CaseInsensitive {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("localfs: invalid grep pattern: %w", err)
	}

	scope := d.root()
	if req.Path != "" {
		p, err := d.resolve(req.Path)
		if err != nil {
			return nil, err
		}
		scope = p
	}

	fileType := strings.TrimPrefix(strings.ToLower(req.FileType), ".")
	matches := make([]filesystem.GrepMatch, 0, 32)
	scanned := 0

	walkErr := filepath.WalkDir(scope, func(p string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || !entry.Type().IsRegular() {
			return nil // 目录、符号链接、设备文件都跳过
		}
		if scanned >= maxGrepFiles || len(matches) >= maxGrepMatches {
			return fs.SkipAll
		}

		rel := d.rel(p)
		if req.Glob != "" {
			if ok, _ := doublestar.Match(req.Glob, rel); !ok {
				return nil
			}
		}
		if fileType != "" && !strings.EqualFold(filepath.Ext(p), "."+fileType) {
			return nil
		}

		info, infoErr := entry.Info()
		if infoErr != nil || info.Size() > maxReadSize {
			return nil
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil || bytes.IndexByte(data, 0) >= 0 {
			return nil // 读失败或含 NUL，按二进制跳过
		}

		scanned++
		content := string(data)
		if req.EnableMultiline {
			matches = appendMultiline(matches, rel, content, re)
		} else {
			matches = appendLineMatches(matches, rel, content, re, req.BeforeLines, req.AfterLines)
		}

		return nil
	})
	if walkErr != nil && !errors.Is(walkErr, fs.SkipAll) {
		return nil, fmt.Errorf("localfs: grep %s: %w", req.Pattern, walkErr)
	}

	return matches, nil
}

// Write 只读实现，不允许 agent 改文件。
func (d Dir) Write(context.Context, *filesystem.WriteRequest) error {
	return os.ErrPermission
}

// Edit 只读实现，不允许 agent 改文件。
func (d Dir) Edit(context.Context, *filesystem.EditRequest) error {
	return os.ErrPermission
}

func appendLineMatches(dst []filesystem.GrepMatch, filePath, content string, re *regexp.Regexp, before, after int) []filesystem.GrepMatch {
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		if !re.MatchString(line) {
			continue
		}
		if before <= 0 && after <= 0 {
			dst = append(dst, filesystem.GrepMatch{Content: line, Path: filePath, Line: idx + 1})
			continue
		}

		start, end := max(0, idx-before), min(len(lines), idx+after+1)
		dst = append(dst, filesystem.GrepMatch{
			Content: strings.Join(lines[start:end], "\n"),
			Path:    filePath,
			Line:    start + 1,
		})
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
