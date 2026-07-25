package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"goraven/backend/infra"
	"goraven/backend/repository/seed/mock"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/core/agent"

	"github.com/8treenet/freedom"
	iris "github.com/8treenet/iris/v12"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/chat", &ChatController{}, infra.NewAuth(true))
	})
}

type ChatController struct {
	ChatSev    *service.ChatService
	SessionSev *service.SessionService
	HfsSev     *service.HFSService
	McpSev     *service.McpService
	SkillSev   *service.SkillService
	Request    *infra.Request
}

const (
	chatUseMock = false
)

func (controller *ChatController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("POST", "/", "Chat")
	b.Handle("POST", "/stop", "Stop")
	b.Handle("GET", "/{sessionId:string}/stream", "Stream")

	b.Handle("POST", "/compress", "Compress")
	b.Handle("GET", "/compress/{taskId:string}", "PollCompress")
}

func (controller *ChatController) Chat() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.ChatReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if chatUseMock {
		mock.MarkSessionCreated(userId, mock.ChatMockSessionId)
		return &infra.JSONResponse{Object: &vo.ChatRsp{SessionId: mock.ChatMockSessionId}}
	}

	rsp, err := controller.ChatSev.StartChat(
		context.Background(),
		userId,
		req,
		controller.McpSev,
		controller.SkillSev,
	)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *ChatController) Stream(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()

	if chatUseMock {
		_ = userId
		return &sseResponse{
			SessionId: sessionId,
			SSEChan:   mock.BuildStreamMock(sessionId),
			Runner:    nil,
		}
	}

	runner, err := controller.ChatSev.GetRunner(sessionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	sseChan := runner.StartFetch()
	return &sseResponse{
		SessionId: sessionId,
		SSEChan:   sseChan,
		Runner:    runner,
	}
}

func (controller *ChatController) Stop() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.ChatStopReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ChatSev.StopChat(req.SessionId, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

func (controller *ChatController) Compress() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.ChatCompressReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.ChatSev.CompressChat(req.SessionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *ChatController) PollCompress(taskId string) freedom.Result {
	rsp, err := controller.ChatSev.PollCompress(taskId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

type sseResponse struct {
	SessionId string
	SSEChan   <-chan *agent.SSEEvent
	Runner    *agent.MainRunner
}

func (sse *sseResponse) Dispatch(ctx iris.Context) {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.StatusCode(200)

	sessionData, _ := json.Marshal(map[string]string{"sessionId": sse.SessionId})
	if _, err := fmt.Fprintf(ctx.ResponseWriter(), "event: connected\ndata: %s\n\n", sessionData); err != nil {
		return
	}
	ctx.ResponseWriter().Flush()

	reqCtx := ctx.Request().Context()

	for {
		select {
		case <-reqCtx.Done():
			if sse.Runner != nil {
				sse.Runner.StopFetchStatus()
			}
			return
		case event, ok := <-sse.SSEChan:
			if !ok {

				return
			}

			data, _ := json.Marshal(event)
			if _, err := fmt.Fprintf(ctx.ResponseWriter(), "event: %s\ndata: %s\n\n", event.Type, data); err != nil {
				return
			}
			ctx.ResponseWriter().Flush()

			if event.Type == agent.SSEEventTypeEnd {
				return
			}
		}
	}
}
