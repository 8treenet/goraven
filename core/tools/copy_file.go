package tools

import (
	"context"

	"goraven/config"
	"goraven/core/sandbox"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

const (
	CopyFileToolDesc = `Copies a file from source path to destination path.

Usage:
- Both source and destination parameters must be absolute paths within the user's workspace
- Destination parent directories will be automatically created if they do not exist
- Returns a confirmation message with the source and destination paths
- Only regular files can be copied (not directories)`

	CopyFileToolDescChinese = `将文件从源路径复制到目标路径。

使用方法：
- source 和 destination 参数都必须是用户工作空间内的绝对路径
- 如果目标父目录不存在，将自动创建
- 返回包含源路径和目标路径的确认信息
- 只能复制常规文件（不能复制目录）`
)

type CopyFileRequest struct {
	Source      string `json:"source" jsonschema:"description=The absolute path of the source file"`
	Destination string `json:"destination" jsonschema:"description=The absolute path of the destination file"`
}

type copyFileTool struct {
	Name    string
	Desc    string
	Sandbox sandbox.Sandbox
}

func NewCopyFile(sandbox sandbox.Sandbox) (tool.InvokableTool, error) {
	desc := CopyFileToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = CopyFileToolDescChinese
	}

	t := &copyFileTool{
		Name:    "cp",
		Desc:    desc,
		Sandbox: sandbox,
	}

	return utils.InferTool(t.Name, t.Desc, t.Invoke)
}

func (t *copyFileTool) Invoke(ctx context.Context, req *CopyFileRequest) (string, error) {
	return t.Sandbox.CopyFile(req.Source, req.Destination)
}
