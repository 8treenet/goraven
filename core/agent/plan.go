package agent

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"sync"

	fspkg "github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/adk/middlewares/plantask"
)

type inMemoryBackend struct {
	files map[string]string
	mu    sync.RWMutex
}

func newInMemoryBackend() *inMemoryBackend {
	return &inMemoryBackend{
		files: make(map[string]string),
	}
}

func (b *inMemoryBackend) LsInfo(ctx context.Context, req *plantask.LsInfoRequest) ([]plantask.FileInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	reqPath := strings.TrimSuffix(req.Path, "/")
	var result []plantask.FileInfo
	for path := range b.files {
		dir := filepath.Dir(path)
		if dir == reqPath {
			result = append(result, plantask.FileInfo{Path: path})
		}
	}
	return result, nil
}

func (b *inMemoryBackend) Read(ctx context.Context, req *plantask.ReadRequest) (*fspkg.FileContent, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	content, ok := b.files[req.FilePath]
	if !ok {
		return nil, errors.New("file not found")
	}
	return &fspkg.FileContent{Content: content}, nil
}

func (b *inMemoryBackend) Write(ctx context.Context, req *plantask.WriteRequest) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.files[req.FilePath] = req.Content
	return nil
}

func (b *inMemoryBackend) Delete(ctx context.Context, req *plantask.DeleteRequest) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.files, req.FilePath)
	return nil
}

// lookupTaskSubject 在内存 backend 中按 <baseDir>/<taskID>.json 读取任务，
// 反序列化后返回 subject。任务不存在或解析失败返回 ("", false)。
// 供 SSE 工具事件展示层（TaskUpdate 解析器）把 taskId 映射成可读 subject。
func (b *inMemoryBackend) lookupTaskSubject(baseDir, taskID string) (string, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	path := filepath.Join(baseDir, taskID+".json")
	content, ok := b.files[path]
	if !ok || content == "" {
		return "", false
	}
	var t struct {
		Subject string `json:"subject"`
	}
	if err := json.Unmarshal([]byte(content), &t); err != nil {
		return "", false
	}
	if t.Subject == "" {
		return "", false
	}
	return t.Subject, true
}
