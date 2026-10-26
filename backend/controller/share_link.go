package controller

import (
	"goraven/backend/infra"
	"goraven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/share", &ShareController{}, infra.NewAuth(false, "/internalMessages"))
	})
}

// ShareController 分享控制器
// GET /{shareId:string}                 无需鉴权
// GET /{shareId:string}/messages        无需鉴权（公开分享消息）
// GET /{shareId:string}/internalMessages 需要鉴权（内部分享消息）
type ShareController struct {
	ShareLinkSev *service.ShareLinkService
	Request      *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/share
func (controller *ShareController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/{shareId:string}", "GetShareInfo")
	b.Handle("GET", "/{shareId:string}/messages", "GetShareMessages")
	b.Handle("GET", "/{shareId:string}/internalMessages", "GetInternalShareMessages")
}

// GetShareInfo 获取分享信息 GET /api/share/:shareId
func (controller *ShareController) GetShareInfo(shareId string) freedom.Result {
	rsp, err := controller.ShareLinkSev.GetShareInfo(shareId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetShareMessages 获取公开分享的消息 GET /api/share/:shareId/messages
func (controller *ShareController) GetShareMessages(shareId string) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.ShareLinkSev.GetShareMessages(shareId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetInternalShareMessages 获取内部分享的消息 GET /api/share/:shareId/internalMessages
func (controller *ShareController) GetInternalShareMessages(shareId string) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.ShareLinkSev.GetShareMessages(shareId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}
