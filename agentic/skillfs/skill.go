// Package skillfs 把技能（SKILL.md + 脚本等）以 zip 的目录层级存进 MongoDB，
// 在库里模拟一台文件系统，并在这台 FS 之上提供 eino 的 skill.Backend。
//
// 一台 FS 对应一个集合（默认 agent_skill_node），每个节点是一条文档：
// 目录是目录、文件是文件，多级嵌套原样保留。文本内联在 content 里（便于 grep），
// 二进制落 GridFS。导入一个 zip 即可把整棵技能树铺进去：
//
//	fs := backend.FS()
//	fs.ImportZip(ctx, "/skills", zipBytes)   // 层级与 zip 完全一致
//
// 技能目录约定为 <BaseDir>/<技能名>/SKILL.md，BaseDir 默认 /skills，因此
// skill.Skill.BaseDirectory 是这台 FS 里的真实路径，可以直接交给文件工具使用。
//
// # 模型白名单
//
// frontmatter 的 model 字段会让技能直接指定模型（ADK 会在 WrapModel 中把整轮
// run 的模型换掉），因此这里默认不允许：未配置 ModelGuard 时，声明了 model 的
// 技能在 List 中不下发该字段、在 Get 中直接报错，写入时也会被拒绝。
package skillfs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultCollection 是默认的虚拟文件系统集合名。
	DefaultCollection = "agent_skill_node"
	// DefaultBlobBucket 是默认的 GridFS 桶名，存放二进制文件内容。
	DefaultBlobBucket = "agent_skill_blob"
	// DefaultBaseDir 是技能目录的默认根，技能位于 <BaseDir>/<name>/SKILL.md。
	DefaultBaseDir = "/skills"
	// SkillFileName 是技能定义文件名。
	SkillFileName = "SKILL.md"
)

var (
	// ErrSkillNotFound 表示指定的技能目录不存在。
	ErrSkillNotFound = errors.New("skillfs: skill not found")
	// ErrModelOverrideDisabled 表示未开启技能级模型覆写，而技能声明了 model。
	ErrModelOverrideDisabled = errors.New("skillfs: model override is disabled")
	// ErrModelNotAllowed 表示技能声明的 model 未通过白名单校验。
	ErrModelNotAllowed = errors.New("skillfs: model not allowed")
)

// ModelGuard 校验技能声明的 model 是否允许使用，返回非 nil 表示拒绝。
// 典型实现直接复用 service.AIChat.LLM，把白名单收敛到一处。
type ModelGuard func(ctx context.Context, name string) error

// Config 是 Backend 的构造参数。
type Config struct {
	// DB 是 Mongo 数据库，必填。
	DB *mongo.Database
	// Collection 是虚拟文件系统节点集合名，默认 DefaultCollection。
	Collection string
	// BlobBucket 是 GridFS 桶名，默认 DefaultBlobBucket。
	BlobBucket string
	// BaseDir 是技能目录根，默认 /skills。
	BaseDir string
	// ModelGuard 开启并校验技能级模型覆写，不配置则禁止技能声明 model。
	ModelGuard ModelGuard
}

// Backend 是 skill.Backend 的 MongoDB 实现，可直接传给 skill.TypedConfig.Backend。
type Backend struct {
	fs      *FS
	baseDir string
	guard   ModelGuard
}

var _ skill.Backend = (*Backend)(nil)

// New 创建基于虚拟文件系统的 skill.Backend。
func New(cfg Config) (*Backend, error) {
	if cfg.DB == nil {
		return nil, errors.New("skillfs: mongo database is required")
	}

	coll := strings.TrimSpace(cfg.Collection)
	if coll == "" {
		coll = DefaultCollection
	}
	bucket := strings.TrimSpace(cfg.BlobBucket)
	if bucket == "" {
		bucket = DefaultBlobBucket
	}
	baseDir := cfg.BaseDir
	if strings.TrimSpace(baseDir) == "" {
		baseDir = DefaultBaseDir
	}

	return &Backend{
		fs:      newFS(cfg.DB, coll, bucket),
		baseDir: normalize(baseDir),
		guard:   cfg.ModelGuard,
	}, nil
}

// FS 返回底层虚拟文件系统，可交给 agent 的文件工具或管理端使用。
func (b *Backend) FS() *FS { return b.fs }

// BaseDir 返回技能目录根。
func (b *Backend) BaseDir() string { return b.baseDir }

// EnsureIndex 创建虚拟文件系统的索引，启动时调用一次即可。
func (b *Backend) EnsureIndex(ctx context.Context) error { return b.fs.EnsureIndex(ctx) }

// List 返回全部技能的元数据，按技能名排序。只解析 SKILL.md，不碰脚本文件。
func (b *Backend) List(ctx context.Context) ([]skill.FrontMatter, error) {
	entries, err := b.fs.GlobInfo(ctx, &filesystem.GlobInfoRequest{
		Pattern: "*/" + SkillFileName,
		Path:    b.baseDir,
	})
	if err != nil {
		return nil, fmt.Errorf("skillfs: list skills: %w", err)
	}

	matters := make([]skill.FrontMatter, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir {
			continue
		}
		fm, _, loadErr := b.load(ctx, entry.Path)
		if loadErr != nil {
			return nil, loadErr
		}
		// 工具描述模板只用 name/description，模型选择不该出现在清单里。
		fm.Model = ""
		matters = append(matters, fm)
	}

	return matters, nil
}

// Get 按技能名读取 SKILL.md，返回带正文与目录的技能定义。
func (b *Backend) Get(ctx context.Context, name string) (skill.Skill, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || strings.ContainsAny(name, "/\\") {
		return skill.Skill{}, fmt.Errorf("skillfs: invalid skill name %q", name)
	}

	dir := path.Join(b.baseDir, name)
	fm, body, err := b.load(ctx, path.Join(dir, SkillFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return skill.Skill{}, fmt.Errorf("%w: %s", ErrSkillNotFound, name)
		}
		return skill.Skill{}, err
	}
	if err = b.allowModel(ctx, fm.Model); err != nil {
		return skill.Skill{}, err
	}

	return skill.Skill{
		FrontMatter:   fm,
		Content:       body,
		BaseDirectory: dir,
	}, nil
}

// PutSkill 把技能写回虚拟文件系统（<BaseDir>/<name>/SKILL.md），
// 便于管理端在线编辑；导入 zip 的场景直接用 FS.ImportZip。
func (b *Backend) PutSkill(ctx context.Context, fm skill.FrontMatter, body string) error {
	name := strings.TrimSpace(fm.Name)
	if name == "" || strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("skillfs: invalid skill name %q", fm.Name)
	}
	if strings.TrimSpace(fm.Description) == "" {
		return errors.New("skillfs: skill description is required")
	}
	if strings.TrimSpace(body) == "" {
		return errors.New("skillfs: skill content is required")
	}
	switch fm.Context {
	case "", skill.ContextModeFork, skill.ContextModeForkWithContext:
	default:
		return fmt.Errorf("skillfs: invalid context %q", fm.Context)
	}
	if err := b.allowModel(ctx, fm.Model); err != nil {
		return err
	}

	content, err := renderSkillFile(fm, body)
	if err != nil {
		return err
	}

	return b.fs.Write(ctx, &filesystem.WriteRequest{
		FilePath: path.Join(b.baseDir, name, SkillFileName),
		Content:  content,
	})
}

// Delete 删除整个技能目录，脚本等附属文件一并清理。
func (b *Backend) Delete(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("skillfs: invalid skill name %q", name)
	}

	if err := b.fs.Remove(ctx, path.Join(b.baseDir, name)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrSkillNotFound, name)
		}
		return err
	}

	return nil
}

func (b *Backend) load(ctx context.Context, filePath string) (skill.FrontMatter, string, error) {
	res, err := b.fs.Read(ctx, &filesystem.ReadRequest{FilePath: filePath})
	if err != nil {
		return skill.FrontMatter{}, "", err
	}

	return parseSkillFile(filePath, res.Content)
}

// allowModel 是技能级模型覆写的唯一闸门：未配置 ModelGuard 时一律拒绝，
// 配置了则交给 guard（例如 service.AIChat.LLM）判定是否在白名单内。
func (b *Backend) allowModel(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}
	if b.guard == nil {
		return fmt.Errorf("%w: %s", ErrModelOverrideDisabled, name)
	}
	if err := b.guard(ctx, name); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrModelNotAllowed, name, err)
	}

	return nil
}

// frontMatterDoc 是 SKILL.md frontmatter 的序列化形态，零值字段直接省略。
type frontMatterDoc struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Context     skill.ContextMode `yaml:"context,omitempty"`
	Agent       string            `yaml:"agent,omitempty"`
	Model       string            `yaml:"model,omitempty"`
}

func renderSkillFile(fm skill.FrontMatter, body string) (string, error) {
	head, err := yaml.Marshal(frontMatterDoc{
		Name:        fm.Name,
		Description: fm.Description,
		Context:     fm.Context,
		Agent:       fm.Agent,
		Model:       fm.Model,
	})
	if err != nil {
		return "", fmt.Errorf("skillfs: marshal frontmatter: %w", err)
	}

	return "---\n" + string(head) + "---\n\n" + strings.TrimSpace(body) + "\n", nil
}

// parseSkillFile 拆分 SKILL.md 的 frontmatter 与正文。
func parseSkillFile(filePath, data string) (skill.FrontMatter, string, error) {
	const delimiter = "---"

	data = strings.TrimSpace(data)
	if !strings.HasPrefix(data, delimiter) {
		return skill.FrontMatter{}, "", fmt.Errorf("skillfs: %s: missing frontmatter delimiter", filePath)
	}

	rest := data[len(delimiter):]
	end := strings.Index(rest, "\n"+delimiter)
	if end < 0 {
		return skill.FrontMatter{}, "", fmt.Errorf("skillfs: %s: unterminated frontmatter", filePath)
	}

	var fm skill.FrontMatter
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return skill.FrontMatter{}, "", fmt.Errorf("skillfs: %s: parse frontmatter: %w", filePath, err)
	}
	if strings.TrimSpace(fm.Name) == "" {
		return skill.FrontMatter{}, "", fmt.Errorf("skillfs: %s: frontmatter name is required", filePath)
	}

	body := strings.TrimSpace(strings.TrimPrefix(rest[end+len(delimiter):], "\n"))

	return fm, body, nil
}
