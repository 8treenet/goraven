package sandbox

import (
	"goraven/core/sandbox/local"
	"goraven/core/sandbox/types"
)

// 类型别名，将 types 包中的类型重新导出到 sandbox 包，保持外部代码兼容
type (
	Sandbox           = types.Sandbox
	StorageUsage      = types.StorageUsage
	StorageCapacity   = types.StorageCapacity
	Backend           = types.Backend
	ShellConfig       = types.ShellConfig
	FileManager       = types.FileManager
	FileManagerConfig = types.FileManagerConfig
	FileInfo          = types.FileInfo
)

// NewSandbox 创建沙盒
func NewSandbox(userName string) (Sandbox, error) {
	return local.NewLocalSandbox(userName), nil
}
