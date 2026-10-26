package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// rootAbsPath 将根目录归一化为清理后的绝对路径。
func rootAbsPath(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root: %w", err)
	}
	return filepath.Clean(abs), nil
}

// openFileRoot 打开以 rootDir 为根的稳定文件系统句柄。
func openFileRoot(rootDir string) (*os.Root, error) {
	rt, err := os.OpenRoot(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open root directory: %w", err)
	}
	return rt, nil
}

// cleanRootRel 词法校验并归一化根目录内的相对路径。
// 空路径、"/"、"." 均归一化为 "."（根目录自身）；逃逸根目录的 ".." 在此拒绝。
func cleanRootRel(rel string) (string, error) {
	if rel == "/" {
		rel = ""
	}
	clean := filepath.Clean(filepath.Join(".", rel))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %s", rel)
	}
	return clean, nil
}

// validateRootPath 将根目录相对路径解析为绝对路径，并检查词法和符号链接边界。
func validateRootPath(root, relPath string) (string, error) {
	workspace, err := rootAbsPath(root)
	if err != nil {
		return "", err
	}
	// 与 cleanRootRel 一致：先归一化根相对路径（"/" 与 "/foo" 等价于 "" 与 "foo"），
	// 归一化后仅剩相对路径，再做逃逸检查，避免把根相对写法误判为绝对路径逃逸。
	relPath = filepath.Clean(filepath.Join(".", relPath))
	if filepath.IsAbs(relPath) || filepath.VolumeName(relPath) != "" {
		return "", fmt.Errorf("path escapes root: %s", relPath)
	}

	cleanRelPath := filepath.Clean(relPath)
	if cleanRelPath == ".." || strings.HasPrefix(cleanRelPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %s", relPath)
	}
	cleanPath := filepath.Join(workspace, cleanRelPath)

	resolvedWorkspace, err := resolveRootExistingPath(workspace)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root: %w", err)
	}
	resolvedPath, err := resolveRootExistingPath(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	if !isWithinRel(resolvedWorkspace, resolvedPath) {
		return "", fmt.Errorf("path escapes root: %s", relPath)
	}
	return cleanPath, nil
}

// resolveRootExistingPath resolves existing symlinks while preserving a non-existent suffix.
func resolveRootExistingPath(path string) (string, error) {
	path = filepath.Clean(path)
	missing := make([]string, 0)

	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		if info, lstatErr := os.Lstat(path); lstatErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("cannot resolve symlink: %s", path)
		}

		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		missing = append(missing, filepath.Base(path))
		path = parent
	}
}

// isWithinRel 判断 child 是否位于 parent 之内（含 parent 本身）。
func isWithinRel(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// IsRootEquivalentPath 判断路径是否等价于根目录本身（空路径、"."、"foo/.." 等形式）。
func IsRootEquivalentPath(path string) bool {
	return filepath.Clean(strings.TrimSpace(path)) == "."
}

// ValidateDirName 校验单个目录名是否符合 Linux 目录命名规则：
// 非空、非 "." 或 ".."、不含 '/' 与 NUL 字节，UTF-8 长度不超过 255 字节。
// 空格、中文、点开头（隐藏目录）、连字符等 Linux 允许的字符均通过。
func ValidateDirName(name string) error {
	if name == "" {
		return fmt.Errorf("directory name is empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid directory name: %s", name)
	}
	if strings.ContainsAny(name, "/\x00") {
		return fmt.Errorf("invalid directory name: %s", name)
	}
	if len(name) > 255 {
		return fmt.Errorf("directory name too long: %d bytes", len(name))
	}
	return nil
}
