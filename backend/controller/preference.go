package controller

import (
	"goraven/backend/infra"
	"goraven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/preference", &PreferenceController{})
	})
}

type PreferenceController struct {
	PrefSev *service.PreferenceService
	Worker  freedom.Worker
	Request *infra.Request
}

func (controller *PreferenceController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "Get")
}

func (controller *PreferenceController) Get() freedom.Result {
	rsp := controller.PrefSev.GetPreference()
	return &infra.JSONResponse{Object: rsp}
}
