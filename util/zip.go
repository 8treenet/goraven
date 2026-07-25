package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func IsSystemZipEntry(name string) bool {

	if strings.HasPrefix(name, "__MACOSX/") {
		return true
	}

	if strings.HasPrefix(name, "._") || strings.Contains(name, "/._") {
		return true
	}

	firstSlash := strings.Index(name, "/")
	if firstSlash == -1 {

		return strings.HasPrefix(name, ".")
	}

	return strings.HasPrefix(name[:firstSlash], ".")
}

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
