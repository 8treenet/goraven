package local

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"goraven/core/sandbox/types"
	"sort"
	"strings"
)

var _ types.FileManager = (*LocalFileManager)(nil)

// LocalFileManager 本地文件管理器
type LocalFileManager struct {
	userName  string
	workspace string
}

// NewLocalFileManager 创建本地文件管理器
func NewLocalFileManager(cfg *types.FileManagerConfig) *LocalFileManager {
	return &LocalFileManager{
		userName:  cfg.UserName,
		workspace: cfg.Workspace,
	}
}

// resolvePath 将相对路径解析为工作空间内的绝对路径，并校验安全性
func (m *LocalFileManager) resolvePath(relPath string) (string, error) {
	cleanWorkspace := filepath.Clean(m.workspace)
	absPath := filepath.Join(cleanWorkspace, relPath)
	cleanPath := filepath.Clean(absPath)

	if !strings.HasPrefix(cleanPath, cleanWorkspace+string(filepath.Separator)) && cleanPath != cleanWorkspace {
		return "", fmt.Errorf("path escapes workspace: %s", relPath)
	}
	return cleanPath, nil
}

// List 列出目录内容
func (m *LocalFileManager) List(dir string, sortBy string, order string) ([]types.FileInfo, error) {
	absDir, err := m.resolvePath(dir)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", dir)
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	dirs := make([]types.FileInfo, 0)
	files := make([]types.FileInfo, 0)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fi := types.FileInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		if entry.IsDir() {
			fi.Size = 0
			dirs = append(dirs, fi)
		} else {
			files = append(files, fi)
		}
	}

	// 排序：文件夹始终在前
	asc := order != "desc"
	sortFunc := fileSortFunc(sortBy, asc)
	sort.Slice(dirs, func(i, j int) bool { return sortFunc(dirs[i], dirs[j]) })
	sort.Slice(files, func(i, j int) bool { return sortFunc(files[i], files[j]) })

	result := make([]types.FileInfo, 0, len(dirs)+len(files))
	result = append(result, dirs...)
	result = append(result, files...)
	return result, nil
}

// ReadFile 读取文件内容
func (m *LocalFileManager) ReadFile(name string) ([]byte, error) {
	absPath, err := m.resolvePath(name)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", name)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", name)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// WriteFile 覆盖写入文件内容（文件不存在则创建，自动创建父目录）
func (m *LocalFileManager) WriteFile(name string, data []byte) error {
	absPath, err := m.resolvePath(name)
	if err != nil {
		return err
	}

	if info, err := os.Stat(absPath); err == nil && info.IsDir() {
		return fmt.Errorf("path is a directory: %s", name)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// Mkdir 创建目录
func (m *LocalFileManager) Mkdir(path string) error {
	absPath, err := m.resolvePath(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// Rename 重命名文件或目录
func (m *LocalFileManager) Rename(oldPath, newPath string) error {
	absOld, err := m.resolvePath(oldPath)
	if err != nil {
		return err
	}
	absNew, err := m.resolvePath(newPath)
	if err != nil {
		return err
	}

	if err := os.Rename(absOld, absNew); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}
	return nil
}

// Delete 删除文件或目录
func (m *LocalFileManager) Delete(paths []string) error {
	for _, path := range paths {
		absPath, err := m.resolvePath(path)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(absPath); err != nil {
			return fmt.Errorf("failed to delete %s: %w", path, err)
		}
	}
	return nil
}

// Compress 压缩为 zip
// 输出文件存放在 paths 中第一个路径所在的目录。
func (m *LocalFileManager) Compress(paths []string, outputName string) (string, error) {
	if len(paths) == 0 {
		return "", fmt.Errorf("paths is empty")
	}
	if !strings.HasSuffix(outputName, ".zip") {
		outputName += ".zip"
	}

	relOutput := filepath.Join(filepath.Dir(paths[0]), outputName)
	absOutput, err := m.resolvePath(relOutput)
	if err != nil {
		return "", err
	}

	zipFile, err := os.Create(absOutput)
	if err != nil {
		return "", fmt.Errorf("failed to create zip: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, path := range paths {
		absPath, err := m.resolvePath(path)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to stat %s: %w", path, err)
		}

		baseName := filepath.Base(absPath)
		if info.IsDir() {
			if err := addDirToZip(zipWriter, absPath, baseName); err != nil {
				return "", err
			}
		} else {
			if err := addFileToZip(zipWriter, absPath, baseName); err != nil {
				return "", err
			}
		}
	}

	return relOutput, nil
}

// Decompress 解压 zip 文件
func (m *LocalFileManager) Decompress(path string, toSubDir bool) error {
	absPath, err := m.resolvePath(path)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(strings.ToLower(absPath), ".zip") {
		return fmt.Errorf("not a zip file: %s", path)
	}

	reader, err := zip.OpenReader(absPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	destDir := filepath.Dir(absPath)
	if toSubDir {
		baseName := strings.TrimSuffix(filepath.Base(absPath), ".zip")
		destDir = filepath.Join(destDir, baseName)
	}

	for _, f := range reader.File {
		destPath := filepath.Join(destDir, f.Name)
		// zip slip 防护
		cleanDest := filepath.Clean(destPath)
		cleanDestDir := filepath.Clean(destDir)
		if !strings.HasPrefix(cleanDest, cleanDestDir+string(filepath.Separator)) && cleanDest != cleanDestDir {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(cleanDest, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanDest), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		srcFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry: %w", err)
		}

		dstFile, err := os.Create(cleanDest)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create file: %w", err)
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	return nil
}

// GetUsage 获取磁盘使用统计
func (m *LocalFileManager) GetUsage() (int64, int, error) {
	var usedSize int64
	var fileCount int

	err := filepath.Walk(m.workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			usedSize += info.Size()
			fileCount++
		}
		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to calculate usage: %w", err)
	}

	return usedSize, fileCount, nil
}

// fileSortFunc 返回排序比较函数
func fileSortFunc(sortBy string, asc bool) func(a, b types.FileInfo) bool {
	multiplier := 1
	if !asc {
		multiplier = -1
	}

	return func(a, b types.FileInfo) bool {
		switch sortBy {
		case "size":
			result := int64(0)
			if a.Size < b.Size {
				result = -1
			} else if a.Size > b.Size {
				result = 1
			}
			return result*int64(multiplier) < 0
		case "time":
			result := a.ModTime.Compare(b.ModTime)
			return result*multiplier < 0
		default: // name
			result := strings.Compare(a.Name, b.Name)
			return result*multiplier < 0
		}
	}
}

// addFileToZip 添加文件到 zip
func addFileToZip(zw *zip.Writer, filePath, zipName string) error {
	src, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipName
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, src)
	return err
}

// addDirToZip 递归添加目录到 zip
func addDirToZip(zw *zip.Writer, dirPath, zipDirName string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(dirPath, path)
		zipName := filepath.Join(zipDirName, relPath)

		if info.IsDir() {
			zipName += "/"
			header := &zip.FileHeader{
				Name:     zipName,
				Method:   zip.Store,
				Modified: info.ModTime(),
			}
			_, err := zw.CreateHeader(header)
			return err
		}

		return addFileToZip(zw, path, zipName)
	})
}
