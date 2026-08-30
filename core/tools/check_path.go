package tools

import (
	"context"
	"fmt"
	"os"

	"goraven/config"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type CheckFileExistsRequest struct {
	Path string `json:"path" jsonschema:"description=Absolute file path to check for existence,required"`
}

type CheckFileExistsResponse struct {
	Exists bool   `json:"exists" jsonschema:"description=Whether the file already exists at this path"`
	Path   string `json:"path" jsonschema:"description=The path that was checked"`
}

type CheckFileExists struct {
	Name string
	Desc string
}

const (
	CheckFileExistsToolDesc = `Check if a file already exists at the given path. Always call this tool before creating a new file to avoid accidentally overwriting existing files. If the path exists, choose a different filename.`

	CheckFileExistsToolDescChinese = `检查文件是否已存在于指定路径。新建文件前必须调用此工具以避免意外覆盖已有文件。若路径已存在，使用不同的文件名。`
)

func NewCheckFileExists() (tool.InvokableTool, error) {
	desc := CheckFileExistsToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = CheckFileExistsToolDescChinese
	}

	t := &CheckFileExists{
		Name: "goraven_check_file_exists",
		Desc: desc,
	}

	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

func (c *CheckFileExists) Invoke(ctx context.Context, req *CheckFileExistsRequest) (*CheckFileExistsResponse, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
	_, err := os.Stat(req.Path)
	exists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check path: %w", err)
	}
	return &CheckFileExistsResponse{
		Exists: exists,
		Path:   req.Path,
	}, nil
}