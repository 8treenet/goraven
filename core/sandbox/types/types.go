package types

import (
	"time"

	"github.com/cloudwego/eino/adk/filesystem"
	mcpclient "github.com/mark3labs/mcp-go/client"
)

// Sandbox 文件系统沙盒接口，用于用户文件操作
type Sandbox interface {
	Exists(filePath string) (bool, error)
	Upload(srcPath string, dstPath string) error
	Download(srcPath string) (string, error)
	NewBackend() (Backend, error)
	NewStreamingShell(cfg *ShellConfig) (StreamingShell, error)
	NewFileManager() FileManager
	GetWorkspace() string
	CopyFile(source, destination string) (string, error)
	DeleteFile(path string) (string, error)
	CreateDirectory(path string) (string, error)
	MoveFile(source, destination string) (string, error)
	NewStdioMCPClient(command string, env []string, args ...string) (*mcpclient.Client, error)
	GetStorageUsage() ([]StorageUsage, error)
	GetStorageCapacity() (StorageCapacity, error)
	// IsSkillInstalled 检查技能依赖是否已安装
	IsSkillInstalled(skillName string) (bool, error)
	// MarkSkillInstalled 标记技能依赖已安装，content 为 LLM 返回内容
	MarkSkillInstalled(skillName string, content string) error
	// Delete 删除用户沙盒（清理用户空间目录）
	Delete() error
	// SetExtraWorkspace 设置额外可访问的工作空间路径，validateFilePath 对此路径下的文件操作放行
	// 典型场景：团队项目位于独立目录，需要允许当前用户读写团队项目目录下的文件
	SetExtraWorkspace(workspace string)
}

// StorageUsage 单个目录的存储使用情况
type StorageUsage struct {
	Name      string // 目录名，如 "documents"
	BytesSize int64  // 目录大小（字节）
}

// StorageCapacity 磁盘容量信息
type StorageCapacity struct {
	TotalBytes int64 // 磁盘总容量（字节）
	FreeBytes  int64 // 磁盘剩余可用空间（字节）
}

type Backend = filesystem.Backend
type StreamingShell = filesystem.StreamingShell

// ShellConfig Shell后端配置
type ShellConfig struct {
	ValidateCommand func(string) error
	// Env 附加环境变量（KEY=VALUE 形式
	Env []string
	// Timeout 单条命令超时时间，0 表示不设超时（依赖外层 ctx）。
	// 仅对前台命令（RunInBackendGround=false）生效。
	Timeout time.Duration
}

// FileInfo 文件/目录信息
type FileInfo struct {
	Name    string    `json:"name"`    // 文件名
	IsDir   bool      `json:"isDir"`   // 是否为目录
	Size    int64     `json:"size"`    // 文件大小（字节），目录为 0
	ModTime time.Time `json:"modTime"` // 修改时间
}

// FileManagerConfig 文件管理器配置
type FileManagerConfig struct {
	UserName  string // 用户名称
	Workspace string // 用户工作空间根路径
}

// FileManager 文件管理器接口
// 所有路径均为相对于用户工作空间的相对路径
type FileManager interface {
	// List 列出目录内容，sortBy 为排序字段（name/size/time），order 为升降序（asc/desc）
	List(dir string, sortBy string, order string) ([]FileInfo, error)
	// ReadFile 读取文件内容
	ReadFile(name string) ([]byte, error)
	// WriteFile 覆盖写入文件内容（文件不存在则创建）
	WriteFile(name string, data []byte) error
	// Mkdir 创建目录（支持多级）
	Mkdir(path string) error
	// Rename 重命名文件或目录
	Rename(oldPath, newPath string) error
	// Delete 删除文件或目录
	Delete(paths []string) error
	// Compress 将指定路径压缩为 zip，返回 zip 文件路径
	Compress(paths []string, outputName string) (string, error)
	// Decompress 解压 zip 文件，toSubDir 为 true 时创建同名子目录
	Decompress(path string, toSubDir bool) error
	// GetUsage 获取磁盘使用统计
	GetUsage() (usedSize int64, fileCount int, err error)
}
