package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"goraven/config"
	"goraven/core/knowledge"
	"goraven/util"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

const (
	DocParseModeRead    = "read"
	DocParseModeConvert = "convert"

	docParseDefaultMaxChars = 50000
	docParseTruncatedMarker = "... [truncated, total "
)

// DocParseRequest 文档解析工具的请求参数
type DocParseRequest struct {
	Mode       string `json:"mode" jsonschema:"description=Operation mode: read extracts document text into the response, convert saves the document as a markdown file,enum=read,enum=convert,required"`
	FilePath   string `json:"file_path" jsonschema:"description=Absolute path of the source document in the sandbox (use the path from the goraven-upload tag for chat attachments), e.g. /path/to/report.pdf,required"`
	OutputPath string `json:"output_path" jsonschema:"description=convert mode only: absolute path for the output markdown file, e.g. /path/to/report.md"`
	Format     string `json:"format" jsonschema:"description=read mode only: output format, markdown (default, preserves headings and tables) or text,enum=markdown,enum=text"`
	MaxChars   int    `json:"max_chars" jsonschema:"description=read mode only: max output characters, content beyond is truncated. Default 50000"`
	Pages      string `json:"pages" jsonschema:"description=read mode only: page filter for PDF, e.g. 1-5 or 1,3,5-8"`
}

// DocParseResponse 文档解析工具的响应
type DocParseResponse struct {
	Content    string `json:"content,omitempty" jsonschema:"description=read mode: extracted document text"`
	Truncated  bool   `json:"truncated,omitempty" jsonschema:"description=read mode: whether the content was truncated"`
	OutputPath string `json:"output_path,omitempty" jsonschema:"description=convert mode: the saved markdown file path"`
}

// DocParse 文档解析工具，通过 docling 脚本读取/转换文档。
// 脚本路径由 paths.scripts 决定（本地 ./scripts，Docker /goraven/scripts）。
type DocParse struct {
	Name           string
	Desc           string
	workspace      string
	extraWorkspace string
}

const (
	DocParseToolDesc = `Read or convert documents (PDF/DOCX/PPTX/XLSX/HTML/CSV/LaTeX/ASCIIDOC) to text or markdown. Mode "read" extracts document text into the response; mode "convert" saves the document as a markdown file. See the goraven-doc-parse skill for usage details.`

	DocParseToolDescChinese = `读取或转换文档（PDF/DOCX/PPTX/XLSX/HTML/CSV/LaTeX/ASCIIDOC）。mode 为 read 时提取文档文本直接返回；mode 为 convert 时将文档转换为 Markdown 文件保存。具体用法见 goraven-doc-parse 技能。`
)

// docParseErrMsg 按系统语言返回错误消息
func docParseErrMsg(zh, en string) string {
	if config.Get().GetLanguage() == "zh" {
		return zh
	}
	return en
}

// normalizeDocParseRequest 校验并规范化请求参数（纯函数，便于测试）
func normalizeDocParseRequest(req *DocParseRequest) (*DocParseRequest, error) {
	if req.Mode != DocParseModeRead && req.Mode != DocParseModeConvert {
		return nil, fmt.Errorf(docParseErrMsg(
			"mode 必须为 read 或 convert，当前值: %s",
			"mode must be read or convert, got: %s"), req.Mode)
	}
	if strings.TrimSpace(req.FilePath) == "" {
		return nil, errors.New(docParseErrMsg(
			"file_path 不能为空",
			"file_path is required"))
	}

	normalized := *req
	normalized.FilePath = filepath.Clean(req.FilePath)

	if normalized.Mode == DocParseModeRead {
		if normalized.Format != "text" {
			normalized.Format = "markdown"
		}
		if normalized.MaxChars <= 0 {
			normalized.MaxChars = docParseDefaultMaxChars
		}
		return &normalized, nil
	}

	// convert 模式校验
	if strings.TrimSpace(normalized.OutputPath) == "" {
		return nil, errors.New(docParseErrMsg(
			"convert 模式必须提供 output_path",
			"convert mode requires output_path"))
	}
	normalized.OutputPath = filepath.Clean(normalized.OutputPath)
	if !strings.EqualFold(filepath.Ext(normalized.OutputPath), ".md") {
		return nil, fmt.Errorf(docParseErrMsg(
			"output_path 必须是 .md 文件，当前值: %s",
			"output_path must be a .md file, got: %s"), normalized.OutputPath)
	}
	if !filepath.IsAbs(normalized.OutputPath) {
		return nil, fmt.Errorf(docParseErrMsg(
			"output_path 必须是绝对路径，当前值: %s",
			"output_path must be an absolute path, got: %s"), normalized.OutputPath)
	}
	if normalized.OutputPath == normalized.FilePath {
		return nil, errors.New(docParseErrMsg(
			"output_path 不能与 file_path 相同",
			"output_path must differ from file_path"))
	}
	return &normalized, nil
}

// NewDocParse 创建文档解析工具
// extraWorkspace 为额外可访问的工作空间根目录（如团队项目目录），为空则仅工作空间可访问。
func NewDocParse(workspace string, extraWorkspace string) (tool.InvokableTool, error) {
	desc := DocParseToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = DocParseToolDescChinese
	}

	t := &DocParse{
		Name:           "goraven_doc_parse",
		Desc:           desc,
		workspace:      workspace,
		extraWorkspace: extraWorkspace,
	}

	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

// roots 返回可访问的根目录列表：工作空间 + 额外工作空间（团队项目）
func (d *DocParse) roots() []string {
	roots := []string{d.workspace}
	if d.extraWorkspace != "" {
		roots = append(roots, d.extraWorkspace)
	}
	return roots
}

// withinRoot 校验路径位于指定根目录内，返回清理后的本地路径
func withinRoot(path, root string) (string, bool) {
	if !filepath.IsAbs(path) {
		return "", false
	}
	cleanPath := filepath.Clean(path)
	cleanRoot := filepath.Clean(root)
	if cleanPath == cleanRoot || strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
		return cleanPath, true
	}
	return "", false
}

// resolveSourcePath 将 agent 提供的路径解析为本地文件路径。
// 优先按原路径（goraven-upload 标签给出的绝对路径）访问，工作空间与团队项目目录均放行；
// 失败时回退为各根路径拼接。
func (d *DocParse) resolveSourcePath(path string) (string, error) {
	for _, root := range d.roots() {
		if local, ok := withinRoot(path, root); ok {
			if info, err := os.Stat(local); err == nil && !info.IsDir() {
				return local, nil
			}
		}
	}
	for _, root := range d.roots() {
		fallback := filepath.Join(root, strings.TrimPrefix(path, "/"))
		if info, err := os.Stat(fallback); err == nil && !info.IsDir() {
			return fallback, nil
		}
	}
	return "", fmt.Errorf("file not found: %s", path)
}

// saveOutput 将本地转换产物写入目标路径。优先按原路径写入；失败时回退为工作空间拼接路径。
func (d *DocParse) saveOutput(localSrc, outputPath string) error {
	if err := copyFileDirect(localSrc, outputPath); err == nil {
		return nil
	}
	return copyFileDirect(localSrc, filepath.Join(d.workspace, strings.TrimPrefix(outputPath, "/")))
}

// copyFileDirect 直接复制单个文件到目标路径（自动创建父目录）
func copyFileDirect(src, dst string) error {
	if !filepath.IsAbs(src) {
		return fmt.Errorf("source path must be absolute: %s", src)
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source not found: %s", src)
		}
		return fmt.Errorf("failed to access source: %w", err)
	}
	if srcInfo.IsDir() {
		return fmt.Errorf("source is a directory, only files can be copied: %s", src)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}
	return nil
}

// Invoke 执行文档解析
func (d *DocParse) Invoke(ctx context.Context, req *DocParseRequest) (*DocParseResponse, error) {
	req, err := normalizeDocParseRequest(req)
	if err != nil {
		return nil, err
	}

	// 工作空间内绝对路径 → 本地文件
	localSrc, err := d.resolveSourcePath(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	if req.Mode == DocParseModeRead {
		content, err := knowledge.ReadFile(ctx, localSrc, knowledge.ReadOptions{
			Format:   req.Format,
			MaxChars: req.MaxChars,
			Pages:    req.Pages,
		})
		if err != nil {
			return nil, err
		}
		return &DocParseResponse{
			Content:   content,
			Truncated: strings.Contains(content, docParseTruncatedMarker),
		}, nil
	}

	// convert：先落到本地临时目录，再写入目标路径
	localDst := filepath.Join(os.TempDir(), "goraven-docparse", util.UUID()+".md")
	defer os.Remove(localDst)
	if err := knowledge.ConvertFile(ctx, localSrc, localDst); err != nil {
		return nil, err
	}
	if err := d.saveOutput(localDst, req.OutputPath); err != nil {
		return nil, fmt.Errorf("failed to save output: %w", err)
	}
	return &DocParseResponse{OutputPath: req.OutputPath}, nil
}
