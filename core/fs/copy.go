package fs

import (
	"fmt"
	"io"
	"os"
)

// CopyFileInto 将临时文件复制为根目录内的目标文件，成功后删除源文件。
// 目标文件通过 os.Root 句柄创建，不受并发符号链接替换影响。
func (r *Root) CopyFileInto(srcPath, relPath string) error {
	return copyFileIntoRoot(srcPath, r.absDir, relPath)
}

// copyFileIntoRoot 将临时文件复制为根目录内的目标文件，成功后删除源文件。
func copyFileIntoRoot(srcPath, rootPath, relPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file: %s", srcPath)
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return fmt.Errorf("failed to open root: %w", err)
	}
	defer root.Close()

	// os.Root 拒绝绝对路径，先归一化根相对写法（"/foo" 等价于 "foo"）。
	relPath, err = cleanRootRel(relPath)
	if err != nil {
		return err
	}
	dst, err := root.OpenFile(relPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Remove(srcPath)
}
