package controller

import (
	"raven/backend/infra"
	"raven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/share", &ShareController{}, infra.NewAuth(false, "/internalMessages"))
	})
}

type ShareController struct {
	ShareLinkSev *service.ShareLinkService
	Request      *infra.Request
}

func (controller *ShareController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/{shareId:string}", "GetShareInfo")
	b.Handle("GET", "/{shareId:string}/messages", "GetShareMessages")
	b.Handle("GET", "/{shareId:string}/internalMessages", "GetInternalShareMessages")
}

func (controller *ShareController) GetShareInfo(shareId string) freedom.Result {
	rsp, err := controller.ShareLinkSev.GetShareInfo(shareId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *ShareController) GetShareMessages(shareId string) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.ShareLinkSev.GetShareMessages(shareId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *ShareController) GetInternalShareMessages(shareId string) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.ShareLinkSev.GetShareMessages(shareId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}
