package types

import (
	"time"

	"github.com/cloudwego/eino/adk/filesystem"
)

type Sandbox interface {
	Exists(filePath string) (bool, error)
	Upload(srcPath string, dstPath string) error
	Download(srcPath string) (string, error)
	NewBackend() (Backend, error)
	NewFileManager() FileManager
	GetWorkspace() string
	DeleteFile(path string) (string, error)
	GetStorageUsage() ([]StorageUsage, error)
	GetStorageCapacity() (StorageCapacity, error)

	IsSkillInstalled(skillName string) (bool, error)

	MarkSkillInstalled(skillName string, content string) error
	Delete() error
}

type StorageUsage struct {
	Name      string
	BytesSize int64
}

type StorageCapacity struct {
	TotalBytes int64
	FreeBytes  int64
}

type Backend = filesystem.Backend

type FileInfo struct {
	Name    string    `json:"name"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

type FileManagerConfig struct {
	UserName  string
	Workspace string
}

type FileManager interface {
	List(dir string, sortBy string, order string) ([]FileInfo, error)

	ReadFile(name string) ([]byte, error)

	WriteFile(name string, data []byte) error

	Mkdir(path string) error

	Rename(oldPath, newPath string) error

	Delete(paths []string) error

	Compress(paths []string, outputName string) (string, error)

	Decompress(path string, toSubDir bool) error

	GetUsage() (usedSize int64, fileCount int, err error)
}
