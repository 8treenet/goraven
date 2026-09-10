package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"goraven/config"
	"goraven/core/sandbox/types"
	"goraven/util/disk"
	"goraven/util/envfile"
	"strings"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	mcpclient "github.com/mark3labs/mcp-go/client"
	gopsutil "github.com/shirou/gopsutil/v4/disk"
)

var _ types.Sandbox = (*LocalSandbox)(nil)

// LocalSandbox 本地文件系统沙盒
type LocalSandbox struct {
	userName         string
	Workspace        string
	extraWorkspace   string // 额外可访问的工作空间（如团队项目目录），validateFilePath 也会对此路径放行
}

// NewLocalSandbox 创建本地沙盒
func NewLocalSandbox(userName string) *LocalSandbox {
	return &LocalSandbox{
		userName:  userName,
		Workspace: config.Get().GetUserSpace(userName),
	}
}

// NewBackend 创建本地文件系统后端，包装为带工作区路径校验的受保护后端
func (s *LocalSandbox) NewBackend() (types.Backend, error) {
	inner, err := local.NewBackend(context.Background(), &local.Config{})
	if err != nil {
		return nil, err
	}
	return newGuardedBackend(inner, s.Workspace), nil
}

// StreamingShell 创建本地Shell
// 会自动读取用户空间根目录下的 .profile，将其中的环境变量与 cfg.Env 合并注入子进程。
// .profile 不存在视为正常情况；解析失败则返回错误以避免静默丢失配置。
func (s *LocalSandbox) NewStreamingShell(cfg *types.ShellConfig) (types.StreamingShell, error) {
	if cfg == nil {
		cfg = &types.ShellConfig{}
	}

	profileEnv, err := s.loadProfileEnv()
	if err != nil {
		return nil, err
	}

	if len(profileEnv) > 0 {
		// 来源优先级：.profile 在前，cfg.Env 在后；同名键由 exec.Cmd 取最后一个，
		// 即调用方显式传入的 cfg.Env 覆盖文件中的同名变量。
		merged := make([]string, 0, len(profileEnv)+len(cfg.Env))
		merged = append(merged, profileEnv...)
		merged = append(merged, cfg.Env...)
		cfg.Env = merged
	}

	return NewLocalShell(cfg), nil
}

// loadProfileEnv 读取用户空间根目录下的 .profile，返回解析后的 KEY=VALUE 列表。
// 文件不存在返回 (nil, nil)。
func (s *LocalSandbox) loadProfileEnv() ([]string, error) {
	data, err := s.NewFileManager().ReadFile(".profile")
	if err != nil {
		// 读取 .profile 失败不阻塞 shell 创建
		return nil, nil
	}

	envs, err := envfile.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse .profile: %w", err)
	}
	return envs, nil
}

// NewFileManager 创建本地文件管理器
func (s *LocalSandbox) NewFileManager() types.FileManager {
	cfg := &types.FileManagerConfig{
		UserName:  s.userName,
		Workspace: s.Workspace,
	}
	return NewLocalFileManager(cfg)
}

// Exists 检查文件是否存在
func (s *LocalSandbox) Exists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !info.IsDir(), nil
}

// Download 从沙盒下载文件到当前系统,srcPath是沙盒内的绝对路径
func (s *LocalSandbox) Download(srcPath string) (string, error) {
	srcCleanPath, err := s.validateFilePath(srcPath)
	if err != nil {
		return "", err
	}
	return srcCleanPath, nil
}

// Upload 复制文件或目录到沙盒,srcPath是当前系统路径，dstPath是沙盒路径
func (s *LocalSandbox) Upload(srcPath string, dstPath string) error {
	if srcPath == "" {
		return fmt.Errorf("source path is required")
	}
	if dstPath == "" {
		return fmt.Errorf("destination path is required")
	}
	if !filepath.IsAbs(srcPath) {
		return fmt.Errorf("source path must be absolute: %s", srcPath)
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source not found: %s", srcPath)
		}
		return fmt.Errorf("failed to access source: %w", err)
	}

	dstCleanPath, err := s.validateFilePath(dstPath)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		if err := copyDirRecursive(srcPath, dstCleanPath); err != nil {
			return fmt.Errorf("failed to upload directory: %w", err)
		}
	} else {
		dstDir := filepath.Dir(dstCleanPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
		if err := copyFileUpload(srcPath, dstCleanPath); err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}
	}

	return nil
}

func (s *LocalSandbox) GetWorkspace() string {
	return s.Workspace
}

// SetExtraWorkspace 设置额外可访问的工作空间路径（如团队项目目录），validateFilePath 对此路径下的文件操作放行
func (s *LocalSandbox) SetExtraWorkspace(workspace string) {
	s.extraWorkspace = workspace
}

func (s *LocalSandbox) CopyFile(source, destination string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("source path is required")
	}
	if destination == "" {
		return "", fmt.Errorf("destination path is required")
	}

	srcPath, err := s.validateFilePath(source)
	if err != nil {
		return "", err
	}

	dstPath, err := s.validateFilePath(destination)
	if err != nil {
		return "", err
	}

	if srcPath == dstPath {
		return "", fmt.Errorf("source and destination are the same: %s", srcPath)
	}
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source file not found: %s", source)
		}
		return "", fmt.Errorf("failed to access source: %w", err)
	}
	if srcInfo.IsDir() {
		return "", fmt.Errorf("source is a directory, only files can be copied: %s", source)
	}

	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return fmt.Sprintf("Copied: %s -> %s", source, destination), nil
}

func (s *LocalSandbox) DeleteFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	cleanPath, err := s.validateFilePath(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file or directory not found: %s", path)
		}
		return "", fmt.Errorf("failed to access path: %w", err)
	}

	if info.IsDir() {
		if err := os.RemoveAll(cleanPath); err != nil {
			return "", fmt.Errorf("failed to delete directory: %w", err)
		}
		return fmt.Sprintf("Directory deleted: %s", path), nil
	}

	if err := os.Remove(cleanPath); err != nil {
		return "", fmt.Errorf("failed to delete file: %w", err)
	}

	return fmt.Sprintf("File deleted: %s", path), nil
}

func (s *LocalSandbox) CreateDirectory(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	cleanPath, err := s.validateFilePath(path)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(cleanPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return fmt.Sprintf("Directory created: %s", path), nil
}

func (s *LocalSandbox) MoveFile(source, destination string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("source path is required")
	}
	if destination == "" {
		return "", fmt.Errorf("destination path is required")
	}

	srcPath, err := s.validateFilePath(source)
	if err != nil {
		return "", err
	}

	dstPath, err := s.validateFilePath(destination)
	if err != nil {
		return "", err
	}

	if srcPath == dstPath {
		return "", fmt.Errorf("source and destination are the same: %s", source)
	}

	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		return "", fmt.Errorf("failed to move: %w", err)
	}

	return fmt.Sprintf("Moved: %s -> %s", source, destination), nil
}

func (s *LocalSandbox) validateFilePath(filePath string) (string, error) {
	if !filepath.IsAbs(filePath) {
		return "", fmt.Errorf("path must be absolute: %s", filePath)
	}

	cleanPath := filepath.Clean(filePath)
	roots := []string{s.GetWorkspace()}
	if s.extraWorkspace != "" {
		roots = append(roots, s.extraWorkspace)
	}
	// 字符串前缀 + 符号链接解析双重校验，防止通过符号链接逃逸工作区
	if err := checkContainment(cleanPath, roots); err != nil {
		return "", err
	}
	return cleanPath, nil
}

func (s *LocalSandbox) NewStdioMCPClient(command string, env []string, args ...string) (*mcpclient.Client, error) {
	return mcpclient.NewStdioMCPClient(command, env, args...)
}

// GetStorageUsage 遍历用户空间下的子目录，统计各目录的存储使用量
func (s *LocalSandbox) GetStorageUsage() ([]types.StorageUsage, error) {
	workspace := s.GetWorkspace()
	entries, err := os.ReadDir(workspace)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	result := make([]types.StorageUsage, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(workspace, entry.Name())
		size := disk.DirSize(dirPath)
		result = append(result, types.StorageUsage{
			Name:      entry.Name(),
			BytesSize: size,
		})
	}
	return result, nil
}

// copyFileUpload 复制文件（保留源权限）
func copyFileUpload(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}

// copyDirRecursive 递归复制目录
func copyDirRecursive(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFileUpload(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete 删除用户沙盒，清理整个用户空间目录
func (s *LocalSandbox) Delete() error {
	workspace := s.GetWorkspace()
	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(workspace); err != nil {
		return fmt.Errorf("failed to delete user workspace: %w", err)
	}
	return nil
}

// IsSkillInstalled 检查技能依赖是否已安装（文件是否存在）
func (s *LocalSandbox) IsSkillInstalled(skillName string) (bool, error) {
	dir := config.Get().GetSkillInstalledDir()
	filePath := filepath.Join(dir, skillName)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// MarkSkillInstalled 标记技能依赖已安装，content 为 LLM 返回内容
func (s *LocalSandbox) MarkSkillInstalled(skillName string, content string) error {
	dir := config.Get().GetSkillInstalledDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create skill installed dir: %w", err)
	}
	filePath := filepath.Join(dir, skillName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write skill installed marker: %w", err)
	}
	return nil
}

// GetStorageCapacity 获取用户空间所在磁盘的总容量和剩余可用空间
// 从系统挂载点中找到工作空间所在的磁盘，返回容量信息
func (s *LocalSandbox) GetStorageCapacity() (types.StorageCapacity, error) {
	workspace := s.GetWorkspace()
	mounts := disk.GetMountPoints()

	var bestMount disk.Info
	for _, m := range mounts {
		if strings.HasPrefix(workspace, m.MountPoint) {
			if bestMount.MountPoint == "" || len(m.MountPoint) > len(bestMount.MountPoint) {
				bestMount = m
			}
		}
	}

	if bestMount.MountPoint != "" {
		return types.StorageCapacity{
			TotalBytes: bestMount.TotalBytes,
			FreeBytes:  bestMount.FreeBytes,
		}, nil
	}

	usage, err := gopsutil.Usage(workspace)
	if err != nil {
		return types.StorageCapacity{}, fmt.Errorf("no mount point found for workspace: %s", workspace)
	}
	return types.StorageCapacity{
		TotalBytes: int64(usage.Total),
		FreeBytes:  int64(usage.Free),
	}, nil
}
