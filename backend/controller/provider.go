package controller

import (
	"strconv"

	"goraven/backend/infra"
	"goraven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/providers", &ProvidersController{}, infra.NewAuth(true))
	})
}

// ProvidersController 用户可用的模型供应商相关接口
type ProvidersController struct {
	ModelSev *service.AIModelService
	Request  *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/providers
func (controller *ProvidersController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/models", "GetModels")
	b.Handle("GET", "/models/{id:string}", "GetModel")
}

// GetModels 获取用户可选模型列表 GET /api/providers/models
func (controller *ProvidersController) GetModels() freedom.Result {
	rsp, err := controller.ModelSev.ListEnabledModels(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetModel 通过模型ID获取单个模型 GET /api/providers/models/:id
func (controller *ProvidersController) GetModel(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.ModelSev.GetEnabledModelByID(modelId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}
