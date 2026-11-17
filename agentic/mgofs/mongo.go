package mgofs

import (
	"context"

	"github.com/cloudwego/eino/adk/filesystem"
)

type Mongo struct {
}

func (m *Mongo) LsInfo(ctx context.Context, req *filesystem.LsInfoRequest) ([]filesystem.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Mongo) Read(ctx context.Context, req *filesystem.ReadRequest) (*filesystem.FileContent, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Mongo) GrepRaw(ctx context.Context, req *filesystem.GrepRequest) ([]filesystem.GrepMatch, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Mongo) GlobInfo(ctx context.Context, req *filesystem.GlobInfoRequest) ([]filesystem.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Mongo) Write(ctx context.Context, req *filesystem.WriteRequest) error {
	//TODO implement me
	panic("implement me")
}

func (m *Mongo) Edit(ctx context.Context, req *filesystem.EditRequest) error {
	//TODO implement me
	panic("implement me")
}
