package controller

import (
	"raven/backend/infra"
	"raven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/share", &ShareController{})
	})
}

type ShareController struct {
	ShareLinkSev	*service.ShareLinkService
	Request		*infra.Request
}

func (controller *ShareController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/{shareId:string}", "GetPublicShare")
}

func (controller *ShareController) GetPublicShare(shareId string) freedom.Result {
	rsp, err := controller.ShareLinkSev.GetPublicShare(shareId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}
