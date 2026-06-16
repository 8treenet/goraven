package controller

import (
	"raven/backend/infra"
	"raven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/mcp", &MCPController{}, infra.NewAuth(true))
	})
}

type MCPController struct {
	McpSev	*service.McpService
	Request	*infra.Request
}

func (controller *MCPController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "GetMCPs")
	b.Handle("GET", "/byIds", "GetMCPsByIDs")
}

func (controller *MCPController) GetMCPs() freedom.Result {
	rsp, err := controller.McpSev.ListEnabledMCPs()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

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
