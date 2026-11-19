package fskill

import (
	"context"

	"github.com/cloudwego/eino/adk/middlewares/skill"
)

type FS struct {
}

func (f *FS) List(ctx context.Context) ([]skill.FrontMatter, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FS) Get(ctx context.Context, name string) (skill.Skill, error) {
	//TODO implement me
	panic("implement me")
}
