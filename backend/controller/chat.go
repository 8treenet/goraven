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

// ChatController 对话聊天接口
type ChatController struct {
	ChatSev    *service.ChatService
	SessionSev *service.SessionService
	HfsSev     *service.HFSService
	SkillSev   *service.SkillService
	Request    *infra.Request
}

// === MOCK开关 === 前端联调期间为 true，注释掉下面两行即可恢复真实 ChatService / Stream 逻辑
const (
	chatUseMock = false
)

// BeforeActivation 绑定路由前缀 /api/chat
func (controller *ChatController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("POST", "/", "Chat")
	b.Handle("POST", "/stop", "Stop")
	b.Handle("GET", "/{sessionId:string}/stream", "Stream")

	b.Handle("POST", "/compress", "Compress")
	b.Handle("GET", "/compress/{taskId:string}", "PollCompress")
}

// Chat 发起对话 POST /api/chat
// 创建或复用会话，启动 Agent 运行器，返回会话 ID
func (controller *ChatController) Chat() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.ChatReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	// === MOCK START === 前端联调期间返回硬编码 sessionId
	if chatUseMock {
		mock.MarkSessionCreated(userId, mock.ChatMockSessionId)
		return &infra.JSONResponse{Object: &vo.ChatRsp{SessionId: mock.ChatMockSessionId}}
	}
	// === MOCK END ===

	rsp, err := controller.ChatSev.StartChat(
		context.Background(),
		userId,
		req,
		controller.SkillSev,
	)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Stream 获取会话的 SSE 流 GET /api/chat/:sessionId/stream
// 前端通过此接口接收 AI Agent 的实时输出事件
func (controller *ChatController) Stream(sessionId string) freedom.Result {
	userId := controller.Request.GetUserId()

	// === MOCK START === 前端联调期间返回 4+ 分钟的脚本化 SSE 流
	if chatUseMock {
		_ = userId
		return &sseResponse{
			SessionId: sessionId,
			SSEChan:   mock.BuildStreamMock(sessionId),
			Runner:    nil,
		}
	}
	// === MOCK END ===

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

// Stop 停止当前生成 POST /api/chat/stop
// 调用运行器的 Terminat 方法中断正在进行的对话
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

// Compress 手动压缩上下文 POST /api/chat/compress
// 返回任务 ID，前端轮询 GET /api/chat/compress/:taskId 获取状态
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

// PollCompress 轮询压缩任务状态 GET /api/chat/compress/:taskId
// 前端每秒轮询，返回任务状态：running/done/failed
func (controller *ChatController) PollCompress(taskId string) freedom.Result {
	rsp, err := controller.ChatSev.PollCompress(taskId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// sseResponse SSE 流式响应，实现 freedom.Result（即 hero.Result）
type sseResponse struct {
	SessionId string                 // 会话ID，首个 session 事件携带
	SSEChan   <-chan *agent.SSEEvent // SSE 事件通道
	Runner    *agent.MainRunner      // 运行器，用于客户端断开时停止 fetch
}

// Dispatch 实现 hero.Result 接口，将 SSE 事件写入 HTTP 响应流
/*
POST /api/chat (无 sessionId)
  → 后端自动创建 session，返回 {sessionId: "abc123"}
  ↓
GET /api/chat/abc123/stream
  → SSE 流建立，首个事件就是 connected
  → event: connected
    data: {"type":"session","sessionId":"abc123"}

	event: reasoning
	data: {"type":"reasoning","content":"正在分析..."}

	event: content
	data: {"type":"content","content":"审查完成..."}

	event: tool
	data: {"type":"tool","name":"read_file","displayName":"文件系统","icon":"📄","action":"正在读取文件"}

	event: retry
	data: {"type":"retry","attempt":1,"maxRetries":3,"error":"connection refused"}

	event: end
	data: {"type":"end"}
*/
func (sse *sseResponse) Dispatch(ctx iris.Context) {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.StatusCode(200)

	// 发送首个 connected 事件
	sessionData, _ := json.Marshal(map[string]string{"sessionId": sse.SessionId})
	if _, err := fmt.Fprintf(ctx.ResponseWriter(), "event: connected\ndata: %s\n\n", sessionData); err != nil {
		return
	}
	ctx.ResponseWriter().Flush()

	// 监听客户端断开
	reqCtx := ctx.Request().Context()

	// 消费 SSE 通道
	for {
		select {
		case <-reqCtx.Done():
			if sse.Runner != nil {
				sse.Runner.StopFetchStatus()
			}
			return
		case event, ok := <-sse.SSEChan:
			if !ok {
				// channel 关闭，正常结束，runner 自清理
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
