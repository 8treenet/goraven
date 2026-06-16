package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsSystemZipEntry 判断是否为系统/隐藏元数据条目
// 包括：__MACOSX/、._*、.DS_Store、以及 . 开头的顶层文件或目录
func IsSystemZipEntry(name string) bool {
	// __MACOSX 目录
	if strings.HasPrefix(name, "__MACOSX/") {
		return true
	}
	// macOS 资源分支文件
	if strings.HasPrefix(name, "._") || strings.Contains(name, "/._") {
		return true
	}
	// 检查首个路径组件是否以 . 开头（隐藏文件/目录）
	firstSlash := strings.Index(name, "/")
	if firstSlash == -1 {
		// 根级文件
		return strings.HasPrefix(name, ".")
	}
	// 根级目录
	return strings.HasPrefix(name[:firstSlash], ".")
}

// ExtractZip 解压 zip 到目标目录（包含 Zip Slip 安全校验）
func ExtractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	for _, f := range reader.File {
		// 跳过 macOS Finder 生成的元数据目录
		if strings.HasPrefix(f.Name, "__MACOSX/") {
			continue
		}

		destPath := filepath.Join(destDir, f.Name)

		cleanDest, err := filepath.Abs(destPath)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		cleanBase, err := filepath.Abs(destDir)
		if err != nil {
			return fmt.Errorf("resolve base: %w", err)
		}
		if !strings.HasPrefix(cleanDest, cleanBase+string(filepath.Separator)) && cleanDest != cleanBase {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("create parent dir: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("create file %s: %w", destPath, err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("extract %s: %w", f.Name, err)
		}
	}

	return nil
}
