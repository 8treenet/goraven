package knowledge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"goraven/config"
)

// doclingTimeout docling 首次调用需加载模型，给足超时时间
const doclingTimeout = 10 * time.Minute

// defaultMaxChars read 模式默认输出字符上限
const defaultMaxChars = 50000

// ReadOptions read 模式选项
type ReadOptions struct {
	Format   string // markdown（默认）或 text
	MaxChars int    // 输出字符上限，<=0 使用 defaultMaxChars
	Pages    string // 页码过滤，如 1-5 或 1,3,5-8，空表示不过滤
}

var pythonNames = []string{"python3", "python"}

func findPython() (string, error) {
	for _, name := range pythonNames {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	if config.Get().GetLanguage() == "zh" {
		return "", errors.New("未找到 python3/python，请安装 Python 后重试")
	}
	return "", errors.New("python3/python not found, please install Python and retry")
}

// doclingScript 解析 docling 脚本路径并检查存在。
// 路径来自 paths.scripts：本地为 ./scripts（相对进程工作目录），Docker 为 /goraven/scripts。
func doclingScript(name string) (string, error) {
	abs, err := filepath.Abs(filepath.Join(config.Get().GetScriptsDir(), "docling", name))
	if err != nil {
		return "", fmt.Errorf("resolve docling script path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("docling script not found: %s", abs)
	}
	return abs, nil
}

// readArgs 构建 read.py 命令行参数（不含 python 本身）
func readArgs(scriptPath, src string, opts ReadOptions) []string {
	format := opts.Format
	if format != "text" {
		format = "markdown"
	}
	maxChars := opts.MaxChars
	if maxChars <= 0 {
		maxChars = defaultMaxChars
	}
	args := []string{scriptPath, "--input", src, "--format", format, "--max-chars", fmt.Sprintf("%d", maxChars)}
	if opts.Pages != "" {
		args = append(args, "--pages", opts.Pages)
	}
	return args
}

// convertArgs 构建 convert.py 命令行参数（不含 python 本身）
func convertArgs(scriptPath, src, dst string) []string {
	return []string{scriptPath, "--input", src, "--output", dst}
}

// doclingDepsError 生成依赖缺失的本地化错误（附安装命令）
func doclingDepsError(stderrText string) error {
	reqPath := filepath.Join(config.Get().GetScriptsDir(), "docling", "requirements.txt")
	if config.Get().GetLanguage() == "zh" {
		return fmt.Errorf("docling 依赖未安装，请执行: uv pip install --no-cache -r %s 后重试（stderr: %s）", reqPath, stderrText)
	}
	return fmt.Errorf("docling is not installed. Run: uv pip install --no-cache -r %s and retry (stderr: %s)", reqPath, stderrText)
}

// runDocling 执行 docling 脚本并返回 stdout。
// stderr 捕获用于错误报告与依赖缺失检测（ModuleNotFoundError）。
// ctx 取消或超时会终止脚本进程。
func runDocling(ctx context.Context, args []string) (string, error) {
	python, err := findPython()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, doclingTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, python, args...)
	// 本机调试（macOS + conda base 环境）时启用：
	// cmdArgs := append([]string{"run", "-n", "base", "python"}, args...)
	// cmd = exec.CommandContext(ctx, "conda", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", fmt.Errorf("docling script interrupted: %w", ctxErr)
		}
		stderrText := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrText, "ModuleNotFoundError") {
			return "", doclingDepsError(stderrText)
		}
		if stderrText == "" {
			stderrText = err.Error()
		}
		return "", fmt.Errorf("docling script failed: %s", stderrText)
	}
	return stdout.String(), nil
}

// ReadFile 通过 docling read.py 提取文档文本（markdown 或纯文本）。
// ctx 取消或超时会终止脚本进程。截断标记由 read.py 追加，调用方可用 strings.Contains(result, "[truncated") 检测。
func ReadFile(ctx context.Context, src string, opts ReadOptions) (string, error) {
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("input file not found: %s", src)
	}
	script, err := doclingScript("read.py")
	if err != nil {
		return "", err
	}
	return runDocling(ctx, readArgs(script, src, opts))
}

// ConvertFile 将源文件（src）通过 docling 转换为 Markdown 文件（dst）。
// ctx 取消或超时会终止脚本进程。src 和 dst 都必须是完整路径带后缀名，如 /users/xxx/hello.pdf 和 /users/xxx/hello.md。
func ConvertFile(ctx context.Context, src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("input file not found: %s", src)
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	script, err := doclingScript("convert.py")
	if err != nil {
		return err
	}

	if _, err := runDocling(ctx, convertArgs(script, src, dst)); err != nil {
		return err
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
	// 本机调试（macOS + conda base 环境）时启用：
	// cmdArgs := append([]string{"run", "-n", "base", "python"}, args...)
	// cmd = exec.Command("conda", cmdArgs...)

	cmd.Stderr = os.Stderr

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("docling chunk failed: %w", err)
	}

	if _, err := os.Stat(dst); err != nil {
		return fmt.Errorf("output file not created: %s", dst)
	}

	return nil
}
