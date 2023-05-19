package profile

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load 加载并序列化配置
func Load(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(v); err != nil {
		return err
	}

	// 读取环境变量
	active := struct {
		Env string `yaml:"env"`
	}{}
	_, _ = file.Seek(0, io.SeekStart)
	if _ = yaml.NewDecoder(file).Decode(&active); active.Env == "" {
		return nil
	}

	ext := filepath.Ext(path)
	env := path[:len(path)-len(ext)] + "-" + active.Env + ext

	cur, err := os.Open(env)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cur.Close()

	return yaml.NewDecoder(cur).Decode(v)
}
