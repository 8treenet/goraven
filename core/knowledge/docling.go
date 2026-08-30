package knowledge

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"goraven/config"
)

var pythonNames = []string{"python3", "python"}

func findPython() (string, error) {
	for _, name := range pythonNames {
		path, err := exec.LookPath(name)
		if err == nil {
			return path, nil
		}
	}
	return "", errors.New("python3/python not found in PATH")
}

// ConvertFile 将源文件（src）通过 docling 转换为 Markdown 文件（dst）。
// src 和 dst 都必须是完整路径带后缀名，如 /users/xxx/hello.txt 和 /users/xxx/hello.md。
func ConvertFile(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("input file not found: %s", src)
	}

	python, err := findPython()
	if err != nil {
		return err
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	scriptPath := filepath.Join(config.Get().Paths.Scripts, "docling", "convert.py")

	args := []string{scriptPath, "--input", src, "--output", dst}

	cmd := exec.Command(python, args...)

	// args = append([]string{"run", "-n", "base", "python"}, args...)
	// cmd = exec.Command("conda", args...)

	cmd.Stderr = os.Stderr

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("docling convert failed: %w", err)
	}

	if _, err := os.Stat(dst); err != nil {
		return fmt.Errorf("output file not created: %s", dst)
	}

	return nil
}

// ChunkFile 将源文档通过 docling + HybridChunker 切割为 chunks，输出到 dst（JSON Lines 格式）。
// src 为完整路径的文档文件，dst 为输出的 .jsonl 文件路径。
// 每行一个 JSON 对象：{"text","heading","page","block_type","chunk_index"}
func ChunkFile(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("input file not found: %s", src)
	}

	python, err := findPython()
	if err != nil {
		return err
	}
	_ = python

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	scriptPath := filepath.Join(config.Get().Paths.Scripts, "docling", "chunk.py")

	args := []string{scriptPath, "--input", src, "--output", dst}

	cmd := exec.Command(python, args...)

	// args = append([]string{"run", "-n", "base", "python"}, args...)
	// cmd = exec.Command("conda", args...)

	cmd.Stderr = os.Stderr

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("docling chunk failed: %w", err)
	}

	if _, err := os.Stat(dst); err != nil {
		return fmt.Errorf("output file not created: %s", dst)
	}

	return nil
}
