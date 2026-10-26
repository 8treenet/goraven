package git

import (
	"context"
	"sync"
)

// cloneRegistry 进程内克隆取消注册表：删除项目时取消进行中的克隆，
// 行删除后 goroutine 不再回写状态。
var cloneRegistry = struct {
	sync.Mutex
	cancels map[string]context.CancelFunc
}{cancels: make(map[string]context.CancelFunc)}

// RegisterCloneCancel 注册克隆取消句柄；同 key 已有句柄先取消。
func RegisterCloneCancel(key string, cancel context.CancelFunc) {
	cloneRegistry.Lock()
	if old, ok := cloneRegistry.cancels[key]; ok {
		old()
	}
	cloneRegistry.cancels[key] = cancel
	cloneRegistry.Unlock()
}

// UnregisterCloneCancel 移除取消句柄（克隆结束清理，不触发取消）。
func UnregisterCloneCancel(key string) {
	cloneRegistry.Lock()
	delete(cloneRegistry.cancels, key)
	cloneRegistry.Unlock()
}

// CancelClone 取消进行中的克隆并移除句柄。
func CancelClone(key string) {
	cloneRegistry.Lock()
	cancel, ok := cloneRegistry.cancels[key]
	if ok {
		delete(cloneRegistry.cancels, key)
	}
	cloneRegistry.Unlock()
	if ok {
		cancel()
	}
}

// CloneCancelRegistered 指定 key 是否存在克隆取消句柄。
func CloneCancelRegistered(key string) bool {
	cloneRegistry.Lock()
	defer cloneRegistry.Unlock()
	_, ok := cloneRegistry.cancels[key]
	return ok
}
