package controller

import (
	"strconv"

	"raven/backend/infra"
	"raven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/providers", &ProvidersController{}, infra.NewAuth(true))
	})
}

type ProvidersController struct {
	ModelSev	*service.AIModelService
	Request		*infra.Request
}

func (controller *ProvidersController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/models", "GetModels")
	b.Handle("GET", "/models/{id:string}", "GetModel")
}

func (controller *ProvidersController) GetModels() freedom.Result {
	rsp, err := controller.ModelSev.ListEnabledModels()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

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
