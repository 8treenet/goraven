package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"goraven/config"
	"goraven/core/sandbox"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

const (
	DeleteFileToolDesc = `Deletes a file or directory at the specified path.

Usage:
- The path parameter must be an absolute path within the user's workspace
- This tool will recursively delete directories and their contents
- Returns a confirmation message with the deleted path
- Use this tool with caution as the deletion is irreversible`

	DeleteFileToolDescChinese = `删除指定路径的文件或目录。

使用方法：
- path 参数必须是用户工作空间内的绝对路径
- 此工具将递归删除目录及其内容
- 返回包含已删除路径的确认信息
- 请谨慎使用此工具，因为删除操作不可逆`
)

type DeleteFileRequest struct {
	Path string `json:"path" jsonschema:"description=The absolute path of the file or directory to delete"`
}

type deleteFileTool struct {
	Name    string
	Desc    string
	Sandbox sandbox.Sandbox
}

func NewDeleteFile(sandbox sandbox.Sandbox) (tool.InvokableTool, error) {
	desc := DeleteFileToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = DeleteFileToolDescChinese
	}

	t := &deleteFileTool{
		Name:    "rm",
		Desc:    desc,
		Sandbox: sandbox,
	}

	return utils.InferTool(t.Name, t.Desc, t.Invoke)
}

func (t *deleteFileTool) Invoke(ctx context.Context, req *DeleteFileRequest) (string, error) {
	if req.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if !filepath.IsAbs(req.Path) {
		return "", fmt.Errorf("path must be absolute: %s", req.Path)
	}
	return t.Sandbox.DeleteFile(req.Path)
}
