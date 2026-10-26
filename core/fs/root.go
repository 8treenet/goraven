// Package fs 提供以指定目录为根的文件系统安全操作。
//
// Root 以目录为根：所有路径均为根目录内相对路径，操作层基于 os.Root 稳定句柄，
// 先词法归一（Clean）再执行，避免"先校验路径后操作"的 TOCTOU 竞态（符号链接替换等）。
// 路径级校验（Resolve）额外解析符号链接，确保解析后的目标不逃逸根目录。
package fs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry 根目录内的目录项。
type Entry struct {
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// Root 以 rootDir 为根的文件系统操作句柄。
type Root struct {
	absDir string
}

// Open 打开以 rootDir 为根的文件系统句柄。
func Open(rootDir string) (*Root, error) {
	absDir, err := rootAbsPath(rootDir)
	if err != nil {
		return nil, err
	}
	return &Root{absDir: absDir}, nil
}

// AbsDir 返回根目录的归一化绝对路径。
func (r *Root) AbsDir() string {
	return r.absDir
}

// Clean 词法校验并归一化根目录内的相对路径。
func (r *Root) Clean(rel string) (string, error) {
	return cleanRootRel(rel)
}

// Resolve 将根目录内相对路径解析为绝对路径，并检查词法和符号链接边界。
func (r *Root) Resolve(relPath string) (string, error) {
	return validateRootPath(r.absDir, relPath)
}

// List 列出根目录内指定目录的条目，目录恒在文件前，按 sortBy/order 排序。
func (r *Root) List(dir, sortBy, order string) ([]Entry, error) {
	rt, err := openFileRoot(r.absDir)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	relDir, err := cleanRootRel(dir)
	if err != nil {
		return nil, err
	}

	dirFile, err := rt.Open(relDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", relDir)
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	entries, err := dirFile.ReadDir(-1)
	closeErr := dirFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to read directory: %w", closeErr)
	}

	dirs := make([]Entry, 0)
	files := make([]Entry, 0)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		re := Entry{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		if entry.IsDir() {
			re.Size = 0
			dirs = append(dirs, re)
		} else {
			files = append(files, re)
		}
	}

	asc := order != "desc"
	compare := func(a, b Entry) int {
		switch sortBy {
		case "size":
			switch {
			case a.Size < b.Size:
				return -1
			case a.Size > b.Size:
				return 1
			}
			return 0
		case "time":
			return a.ModTime.Compare(b.ModTime)
		default: // name
			return strings.Compare(a.Name, b.Name)
		}
	}
	sortEntries := func(list []Entry) {
		sort.Slice(list, func(i, j int) bool {
			c := compare(list[i], list[j])
			if !asc {
				c = -c
			}
			return c < 0
		})
	}
	sortEntries(dirs)
	sortEntries(files)

	result := make([]Entry, 0, len(dirs)+len(files))
	result = append(result, dirs...)
	result = append(result, files...)
	return result, nil
}

// MkdirAll 在根目录内递归创建目录。
func (r *Root) MkdirAll(rel string) error {
	rt, err := openFileRoot(r.absDir)
	if err != nil {
		return err
	}
	defer rt.Close()

	rel, err = cleanRootRel(rel)
	if err != nil {
		return err
	}
	if err := rt.MkdirAll(rel, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// Rename 在根目录内重命名文件或目录，根目录本身不可作为源或目标。
func (r *Root) Rename(oldRel, newRel string) error {
	oldRel, err := cleanRootRel(oldRel)
	if err != nil {
		return err
	}
	newRel, err = cleanRootRel(newRel)
	if err != nil {
		return err
	}
	if oldRel == "." || newRel == "." {
		return fmt.Errorf("refusing to rename the root directory")
	}

	rt, err := openFileRoot(r.absDir)
	if err != nil {
		return err
	}
	defer rt.Close()
	if err := rt.Rename(oldRel, newRel); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}
	return nil
}

// RemoveAll 删除根目录内的文件或目录，拒绝删除根目录自身。
func (r *Root) RemoveAll(rels []string) error {
	rt, err := openFileRoot(r.absDir)
	if err != nil {
		return err
	}
	defer rt.Close()

	for _, rel := range rels {
		rel, err := cleanRootRel(rel)
		if err != nil {
			return err
		}
		if rel == "." {
			return fmt.Errorf("refusing to delete the root directory")
		}
		if err := rt.RemoveAll(rel); err != nil {
			return fmt.Errorf("failed to delete %s: %w", rel, err)
		}
	}
	return nil
}

// Usage 遍历根目录统计已用字节数与文件个数。
func (r *Root) Usage() (usedSize int64, fileCount int, err error) {
	rt, err := openFileRoot(r.absDir)
	if err != nil {
		return 0, 0, err
	}
	defer rt.Close()

	err = fs.WalkDir(rt.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, infoErr := d.Info()
			if infoErr != nil {
				return nil
			}
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

// ReadFile 读取根目录内文件内容。
func (r *Root) ReadFile(rel string) ([]byte, error) {
	rt, err := openFileRoot(r.absDir)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	rel, err = cleanRootRel(rel)
	if err != nil {
		return nil, err
	}
	file, err := rt.Open(rel)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", rel)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// WriteFile 覆盖写入根目录内文件（父目录自动创建），目标为目录时拒绝。
func (r *Root) WriteFile(rel string, data []byte) error {
	rt, err := openFileRoot(r.absDir)
	if err != nil {
		return err
	}
	defer rt.Close()

	rel, err = cleanRootRel(rel)
	if err != nil {
		return err
	}
	if info, statErr := rt.Stat(rel); statErr == nil && info.IsDir() {
		return fmt.Errorf("path is a directory: %s", rel)
	}
	if err := rt.MkdirAll(filepath.Dir(rel), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	if err := rt.WriteFile(rel, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
