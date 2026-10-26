package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Zip 将根目录内指定路径压缩为 zip。
// 输出文件存放在 paths 中第一个路径所在的目录，返回 zip 的根目录相对路径。
func (r *Root) Zip(paths []string, outputName string) (string, error) {
	rootDir := r.absDir
	if len(paths) == 0 {
		return "", fmt.Errorf("paths is empty")
	}
	if !strings.HasSuffix(outputName, ".zip") {
		outputName += ".zip"
	}

	rt, err := openFileRoot(rootDir)
	if err != nil {
		return "", err
	}
	defer rt.Close()

	rootPaths := make([]string, len(paths))
	for i, path := range paths {
		rootPath, err := cleanRootRel(path)
		if err != nil {
			return "", err
		}
		rootPaths[i] = rootPath
		info, err := rt.Stat(rootPath)
		if err != nil {
			return "", fmt.Errorf("failed to stat %s: %w", path, err)
		}
		if info.IsDir() {
			if err := zipValidateSourceTree(rt, rootPath); err != nil {
				return "", fmt.Errorf("failed to validate archive source %s: %w", path, err)
			}
		}
	}
	relOutput := filepath.Join(filepath.Dir(paths[0]), outputName)
	rootOutput, err := cleanRootRel(relOutput)
	if err != nil {
		return "", err
	}

	// 先写入系统临时文件，源遍历完成后再发布到根目录，
	// 避免压缩包把自身输出包含进来（如 sources 为 "." 或包含输出路径时）。
	tmpFile, err := os.CreateTemp("", "goraven-compress-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp archive: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	zipWriter := zip.NewWriter(tmpFile)
	if err := zipWriteSources(rt, zipWriter, rootPaths, rootDir); err != nil {
		zipWriter.Close()
		tmpFile.Close()
		return "", err
	}
	if err := zipWriter.Close(); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to finalize archive: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close archive: %w", err)
	}

	// 源遍历完成后发布到根目录
	src, err := os.Open(tmpName)
	if err != nil {
		return "", fmt.Errorf("failed to open temp archive: %w", err)
	}
	defer src.Close()
	dst, err := rt.Create(rootOutput)
	if err != nil {
		return "", fmt.Errorf("failed to create zip: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		rt.Remove(rootOutput)
		return "", fmt.Errorf("failed to publish archive: %w", err)
	}
	if err := dst.Close(); err != nil {
		rt.Remove(rootOutput)
		return "", fmt.Errorf("failed to finalize zip: %w", err)
	}
	return relOutput, nil
}

// Unzip 解压根目录内的 zip 文件，toSubDir 为 true 时解压到同名子目录。
func (r *Root) Unzip(zipRel string, toSubDir bool) error {
	rootDir := r.absDir
	zipRel, err := cleanRootRel(zipRel)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(strings.ToLower(zipRel), ".zip") {
		return fmt.Errorf("not a zip file: %s", zipRel)
	}

	rt, err := openFileRoot(rootDir)
	if err != nil {
		return err
	}
	defer rt.Close()

	zipFile, err := rt.Open(zipRel)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipFile.Close()
	zipInfo, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat zip: %w", err)
	}
	reader, err := zip.NewReader(zipFile, zipInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to read zip: %w", err)
	}

	destDir := filepath.Dir(zipRel)
	if toSubDir {
		baseName := strings.TrimSuffix(filepath.Base(zipRel), ".zip")
		destDir = filepath.Join(destDir, baseName)
	}

	for _, f := range reader.File {
		entryName := filepath.FromSlash(f.Name)
		if filepath.IsAbs(entryName) || filepath.VolumeName(entryName) != "" {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}
		// zip slip 防护：目标必须严格位于解压目录内
		cleanDest, err := cleanRootRel(filepath.Join(destDir, entryName))
		if err != nil {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}
		cleanDestDir, err := cleanRootRel(destDir)
		if err != nil {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}
		if !isWithinRel(cleanDestDir, cleanDest) {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := rt.MkdirAll(cleanDest, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := rt.MkdirAll(filepath.Dir(cleanDest), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		srcFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry: %w", err)
		}

		dstFile, err := rt.Create(cleanDest)
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

// zipValidateSourceTree 校验归档源目录树内不包含悬空/逃逸的符号链接。
func zipValidateSourceTree(root *os.Root, rel string) error {
	info, err := root.Lstat(rel)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		_, err := root.Stat(rel)
		return err
	}
	if !info.IsDir() {
		return nil
	}

	dir, err := root.Open(rel)
	if err != nil {
		return err
	}
	entries, err := dir.ReadDir(-1)
	closeErr := dir.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}

	for _, entry := range entries {
		if err := zipValidateSourceTree(root, filepath.Join(rel, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// zipWriteSources 将根目录内的各源路径写入 zip。
func zipWriteSources(root *os.Root, zipWriter *zip.Writer, rootPaths []string, rootDir string) error {
	for i, rootPath := range rootPaths {
		info, err := root.Stat(rootPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", rootPaths[i], err)
		}

		baseName := filepath.Base(rootPath)
		if rootPath == "." {
			absRoot, err := filepath.Abs(rootDir)
			if err != nil {
				return fmt.Errorf("failed to resolve root: %w", err)
			}
			baseName = filepath.Base(filepath.Clean(absRoot))
		}
		if info.IsDir() {
			if err := zipAddRootDir(root, zipWriter, rootPath, baseName); err != nil {
				return err
			}
		} else {
			if err := zipAddRootFile(root, zipWriter, rootPath, baseName); err != nil {
				return err
			}
		}
	}
	return nil
}

// zipAddRootFile 将根目录内单个文件写入 zip。
func zipAddRootFile(root *os.Root, zw *zip.Writer, filePath, zipName string) error {
	src, err := root.Open(filePath)
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

// zipAddRootDir 递归将根目录内目录写入 zip。
func zipAddRootDir(root *os.Root, zw *zip.Writer, dirPath, zipDirName string) error {
	info, err := root.Stat(dirPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}

	header := &zip.FileHeader{
		Name:     zipDirName + "/",
		Method:   zip.Store,
		Modified: info.ModTime(),
	}
	if _, err := zw.CreateHeader(header); err != nil {
		return err
	}

	dir, err := root.Open(dirPath)
	if err != nil {
		return err
	}
	entries, err := dir.ReadDir(-1)
	closeErr := dir.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		childPath := filepath.Join(dirPath, entry.Name())
		childZipName := filepath.Join(zipDirName, entry.Name())
		childInfo, err := root.Stat(childPath)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if err := zipAddRootDir(root, zw, childPath, childZipName); err != nil {
				return err
			}
			continue
		}
		if childInfo.IsDir() {
			return fmt.Errorf("cannot archive symlink directory: %s", childPath)
		}
		if err := zipAddRootFile(root, zw, childPath, childZipName); err != nil {
			return err
		}
	}
	return nil
}
