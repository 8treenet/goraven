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

type SessionController struct {
	SessionSev   *service.SessionService
	ShareLinkSev *service.ShareLinkService
	Request      *infra.Request
}

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

func (controller *SessionController) ListSessions() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.SessionListReq{}
	if err := controller.Request.ReadQuery(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if chatUseMock {
		return &infra.JSONResponse{Object: mock.BuildMockSessionList(userId)}
	}

	rsp, err := controller.SessionSev.ListSessions(userId, req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SessionController) GetSession(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()

	if chatUseMock {
		detail := mock.BuildMockSessionDetail(sessionId, userId)
		if detail == nil {
			return &infra.JSONResponse{Error: errs.ErrSessionNotFound}
		}
		return &infra.JSONResponse{Object: detail}
	}

	rsp, err := controller.SessionSev.GetSession(sessionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

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

func (controller *SessionController) DeleteSession(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SessionSev.DeleteSession(sessionId, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

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

func (controller *SessionController) GetSessionShare(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.ShareLinkSev.GetSessionShare(sessionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SessionController) DeleteShare(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.ShareLinkSev.DeleteShare(sessionId, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}
