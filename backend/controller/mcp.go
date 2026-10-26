package controller

import (
	"goraven/backend/infra"
	"goraven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/mcp", &MCPController{}, infra.NewAuth(true))
	})
}

// MCPController 用户可用的 MCP 工具相关接口
type MCPController struct {
	McpSev  *service.McpService
	Request *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/mcp
func (controller *MCPController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "GetMCPs")
	b.Handle("GET", "/byIds", "GetMCPsByIDs")
}

// GetMCPs 获取用户可选 MCP 列表 GET /api/mcp
func (controller *MCPController) GetMCPs() freedom.Result {
	rsp, err := controller.McpSev.ListEnabledMCPs()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetMCPsByIDs 根据指定的 mcpId 列表获取 MCP 列表 GET /api/mcp/byIds?ids=1&ids=2
func (controller *MCPController) GetMCPsByIDs() freedom.Result {
	var query struct {
		IDs []int `url:"ids" validate:"required"`
	}
	if err := controller.Request.ReadQuery(&query, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.McpSev.ListEnabledMCPsByIDs(query.IDs)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}
