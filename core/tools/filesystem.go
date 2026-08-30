package tools

import (
	"context"
	"fmt"

	"goraven/config"
	"goraven/core/sandbox"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

const (
	CreateDirectoryToolDesc = `Creates a new directory in the filesystem, creating any missing parent directories as needed.

Usage:
- The path parameter must be an absolute path within the user's workspace
- This tool will create all intermediate directories if they do not exist (equivalent to mkdir -p)
- Returns a confirmation message with the created directory path
- Use this tool before writing files to ensure target directories exist`

	CreateDirectoryToolDescChinese = `在文件系统中创建新目录，根据需要创建所有缺失的父目录。

使用方法：
- path 参数必须是用户工作空间内的绝对路径
- 此工具将创建所有不存在的中间目录（相当于 mkdir -p）
- 返回包含已创建目录路径的确认信息
- 在写入文件之前使用此工具，确保目标目录存在`

	MoveFileToolDesc = `Moves or renames a file or directory in the filesystem.

Usage:
- Both source and destination parameters must be absolute paths within the user's workspace
- This tool can be used to rename files within the same directory or move files across directories
- Destination parent directories will be automatically created if they do not exist
- Returns a confirmation message with the source and destination paths
- To rename a file, provide a different destination path in the same directory`

	MoveFileToolDescChinese = `在文件系统中移动或重命名文件或目录。

使用方法：
- source 和 destination 参数都必须是用户工作空间内的绝对路径
- 此工具可用于在同一目录内重命名文件，或跨目录移动文件
- 如果目标父目录不存在，将自动创建
- 返回包含源路径和目标路径的确认信息
- 要重命名文件，请在同一目录中提供不同的目标路径`
)

// ─── mkdir ───

type CreateDirectoryRequest struct {
	Path string `json:"path" jsonschema:"description=The absolute path of the directory to create"`
}

type createDirectoryTool struct {
	Name    string
	Desc    string
	Sandbox sandbox.Sandbox
}

func NewCreateDirectory(sandbox sandbox.Sandbox) (tool.InvokableTool, error) {
	desc := CreateDirectoryToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = CreateDirectoryToolDescChinese
	}

	t := &createDirectoryTool{
		Name:    "mkdir",
		Desc:    desc,
		Sandbox: sandbox,
	}

	return utils.InferTool(t.Name, t.Desc, t.Invoke)
}

func (t *createDirectoryTool) Invoke(ctx context.Context, req *CreateDirectoryRequest) (string, error) {
	if req.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	return t.Sandbox.CreateDirectory(req.Path)
}

// ─── mv ───

type MoveFileRequest struct {
	Source      string `json:"source" jsonschema:"description=The absolute path of the source file or directory"`
	Destination string `json:"destination" jsonschema:"description=The absolute path of the destination file or directory"`
}

type moveFileTool struct {
	Name    string
	Desc    string
	Sandbox sandbox.Sandbox
}

func NewMoveFile(sandbox sandbox.Sandbox) (tool.InvokableTool, error) {
	desc := MoveFileToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = MoveFileToolDescChinese
	}

	t := &moveFileTool{
		Name:    "mv",
		Desc:    desc,
		Sandbox: sandbox,
	}

	return utils.InferTool(t.Name, t.Desc, t.Invoke)
}

func (t *moveFileTool) Invoke(ctx context.Context, req *MoveFileRequest) (string, error) {
	if req.Source == "" {
		return "", fmt.Errorf("source path is required")
	}
	if req.Destination == "" {
		return "", fmt.Errorf("destination path is required")
	}

	return t.Sandbox.MoveFile(req.Source, req.Destination)
}
