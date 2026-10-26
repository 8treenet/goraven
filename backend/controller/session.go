package controller

import (
	"goraven/backend/infra"
	"goraven/backend/repository/seed/mock"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/sessions", &SessionController{}, infra.NewAuth(true))
	})
}

// SessionController 用户会话相关接口
type SessionController struct {
	SessionSev   *service.SessionService
	ShareLinkSev *service.ShareLinkService
	Request      *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/sessions
func (controller *SessionController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "ListSessions")
	b.Handle("GET", "/shares", "ListMyShares")
	b.Handle("GET", "/{sessionId:string}", "GetSession")
	b.Handle("PUT", "/{sessionId:string}", "UpdateSession")
	b.Handle("GET", "/{sessionId:string}/messages", "GetMessages")
	b.Handle("DELETE", "/{sessionId:string}", "DeleteSession")
	b.Handle("POST", "/{sessionId:string}/share", "CreateShare")
	b.Handle("GET", "/{sessionId:string}/share", "GetSessionShare")
	b.Handle("DELETE", "/{sessionId:string}/share", "DeleteShare")
}

// ListSessions 获取会话列表（侧边栏"所有对话"） GET /api/sessions
func (controller *SessionController) ListSessions() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.SessionListReq{}
	if err := controller.Request.ReadQuery(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	// === MOCK START === 前端联调期间：未调过 Chat() 返回空，调过则返回单条 mock 会话
	if chatUseMock {
		return &infra.JSONResponse{Object: mock.BuildMockSessionList(userId)}
	}
	// === MOCK END ===

	rsp, err := controller.SessionSev.ListSessions(userId, req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetSession 获取会话详情 GET /api/sessions/:sessionId
func (controller *SessionController) GetSession(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()

	// === MOCK START === 前端联调期间：仅在 sessionId 命中 mock 会话时返回详情
	if chatUseMock {
		detail := mock.BuildMockSessionDetail(sessionId, userId)
		if detail == nil {
			return &infra.JSONResponse{Error: errs.ErrSessionNotFound}
		}
		return &infra.JSONResponse{Object: detail}
	}
	// === MOCK END ===

	rsp, err := controller.SessionSev.GetSession(sessionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// UpdateSession 更新会话（标题、归档等） PUT /api/sessions/:sessionId
func (controller *SessionController) UpdateSession(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.SessionUpdateReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.SessionSev.UpdateSession(sessionId, userId, req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetMessages 获取会话历史消息 GET /api/sessions/:sessionId/messages
func (controller *SessionController) GetMessages(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()

	if chatUseMock {
		_ = userId
		return &infra.JSONResponse{Object: []vo.MessageItem{}}
	}

	rsp, err := controller.SessionSev.GetMessages(sessionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// DeleteSession 删除会话（软删除） DELETE /api/sessions/:sessionId
func (controller *SessionController) DeleteSession(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SessionSev.DeleteSession(sessionId, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ListMyShares 获取用户分享链接分页列表 GET /api/sessions/shares
func (controller *SessionController) ListMyShares() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.UserShareListReq{}
	if err := controller.Request.ReadQuery(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.ShareLinkSev.ListUserShares(userId, req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// CreateShare 创建会话分享链接 POST /api/sessions/:sessionId/share
func (controller *SessionController) CreateShare(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.CreateShareReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.ShareLinkSev.CreateShare(userId, sessionId, req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetSessionShare 获取会话的分享链接 GET /api/sessions/:sessionId/share
func (controller *SessionController) GetSessionShare(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.ShareLinkSev.GetSessionShare(sessionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// DeleteShare 删除会话的分享链接 DELETE /api/sessions/:sessionId/share
func (controller *SessionController) DeleteShare(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.ShareLinkSev.DeleteShare(sessionId, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}
