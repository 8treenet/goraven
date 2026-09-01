package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolveSymlinkPath 解析路径中的符号链接，返回真实路径。
//
// 路径本身不存在时，解析其最深的存在着的父目录后拼回剩余部分，
// 以便对"将要写入的文件"也能提前发现逃逸链接。
func resolveSymlinkPath(p string) (string, error) {
	resolved, err := filepath.EvalSymlinks(p)
	if err == nil {
		return resolved, nil
	}

	// 路径不存在：向上找到最深的存在着的祖先
	dir := filepath.Dir(p)
	rest := filepath.Base(p)
	for {
		if dir == string(filepath.Separator) || dir == "." {
			break
		}
		if _, statErr := os.Lstat(dir); statErr == nil {
			break
		}
		rest = filepath.Join(filepath.Base(dir), rest)
		dir = filepath.Dir(dir)
	}

	resolvedDir, evalErr := filepath.EvalSymlinks(dir)
	if evalErr != nil {
		return "", err
	}
	return filepath.Join(resolvedDir, rest), nil
}

// pathWithin 判断 path 是否位于 base 目录内（或等于 base）。
// 仅做字符串比较，调用方需保证两者都已 Clean。
func pathWithin(path, base string) bool {
	if path == base {
		return true
	}
	return strings.HasPrefix(path, base+string(filepath.Separator))
}

// checkContainment 校验路径（字符串前缀 + 符号链接解析后）都落在允许的根目录内。
// roots 为允许的根目录列表；返回 nil 表示安全，否则返回错误。
func checkContainment(cleanPath string, roots []string) error {
	resolvedRoots := make([]string, 0, len(roots))
	allowed := false
	for _, root := range roots {
		if root == "" {
			continue
		}
		cleanRoot := filepath.Clean(root)
		resolvedRoots = append(resolvedRoots, cleanRoot)
		if resolvedRoot, err := filepath.EvalSymlinks(cleanRoot); err == nil {
			resolvedRoots = append(resolvedRoots, resolvedRoot)
		}
		if pathWithin(cleanPath, cleanRoot) {
			allowed = true
		}
	}
	if !allowed {
		return fmt.Errorf("path is outside allowed roots: %s", cleanPath)
	}

	// 符号链接逃逸防护：解析后的真实路径必须仍位于允许的根目录内
	resolved, err := resolveSymlinkPath(cleanPath)
	if err != nil {
		return err
	}
	for _, root := range resolvedRoots {
		if pathWithin(resolved, root) {
			return nil
		}
	}
	return fmt.Errorf("symlink escapes allowed roots: %s", cleanPath)
}
