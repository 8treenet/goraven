package local

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cloudwego/eino/adk/filesystem"
)

// guardedBackend 在 eino 本地文件系统后端之上增加工作区路径校验。
//
// eino-ext 的 local 后端没有任何根目录约束（默认根为 "/"），Agent 的
// read/write/edit/grep/glob/ls 工具可直接访问宿主文件系统。本装饰器
// 在委托前把所有路径参数约束到用户工作区内（含符号链接逃逸检查），
// 相对路径按工作区根解析。
type guardedBackend struct {
	inner     filesystem.Backend
	workspace string
}

var _ filesystem.Backend = (*guardedBackend)(nil)

// newGuardedBackend 创建带工作区校验的文件系统后端。
func newGuardedBackend(inner filesystem.Backend, workspace string) *guardedBackend {
	return &guardedBackend{inner: inner, workspace: workspace}
}

// validate 将路径约束到工作区内，返回清洗后的绝对路径。
func (g *guardedBackend) validate(p string) (string, error) {
	if p == "" {
		return g.workspace, nil
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(g.workspace, p)
	}
	cleanPath := filepath.Clean(p)
	if err := checkContainment(cleanPath, []string{g.workspace}); err != nil {
		return "", fmt.Errorf("path is outside the agent workspace: %s", p)
	}
	return cleanPath, nil
}

func (g *guardedBackend) LsInfo(ctx context.Context, req *filesystem.LsInfoRequest) ([]filesystem.FileInfo, error) {
	validated, err := g.validate(req.Path)
	if err != nil {
		return nil, err
	}
	req.Path = validated
	return g.inner.LsInfo(ctx, req)
}

func (g *guardedBackend) Read(ctx context.Context, req *filesystem.ReadRequest) (*filesystem.FileContent, error) {
	validated, err := g.validate(req.FilePath)
	if err != nil {
		return nil, err
	}
	req.FilePath = validated
	return g.inner.Read(ctx, req)
}

func (g *guardedBackend) GrepRaw(ctx context.Context, req *filesystem.GrepRequest) ([]filesystem.GrepMatch, error) {
	// Path 为空时 backend 默认从进程根开始搜索，强制收敛到工作区
	validated, err := g.validate(req.Path)
	if err != nil {
		return nil, err
	}
	req.Path = validated
	return g.inner.GrepRaw(ctx, req)
}

func (g *guardedBackend) GlobInfo(ctx context.Context, req *filesystem.GlobInfoRequest) ([]filesystem.FileInfo, error) {
	validated, err := g.validate(req.Path)
	if err != nil {
		return nil, err
	}
	req.Path = validated
	return g.inner.GlobInfo(ctx, req)
}

func (g *guardedBackend) Write(ctx context.Context, req *filesystem.WriteRequest) error {
	validated, err := g.validate(req.FilePath)
	if err != nil {
		return err
	}
	req.FilePath = validated
	return g.inner.Write(ctx, req)
}

func (g *guardedBackend) Edit(ctx context.Context, req *filesystem.EditRequest) error {
	validated, err := g.validate(req.FilePath)
	if err != nil {
		return err
	}
	req.FilePath = validated
	return g.inner.Edit(ctx, req)
}
