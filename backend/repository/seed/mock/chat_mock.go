package mock

import (
	"sync"
	"time"

	"goraven/backend/infra"
	"goraven/backend/vo"
)

// ChatMockSessionId 是 Chat 接口 mock 返回的固定 sessionId。
// 前端拿到该 id 后会立即调用 GET /api/chat/{sessionId}/stream 拉取 SSE 流。
const ChatMockSessionId = "mock-session-2026-demo"

// sessionRegistry 记录“哪些 user 已经调过 Chat、返回的 sessionId 是什么”。
// SessionList / SessionDetail 接口从这里读，按用户隔离。
//
// 使用 sync.Map 而不是普通 map，避免 controller 并发场景下的 data race。
var sessionRegistry sync.Map

// MarkSessionCreated 在 Chat() mock 命中时调用一次，
// 把 (userId, sessionId) 写入 sessionRegistry，供后续的会话列表 / 详情查询。
func MarkSessionCreated(userId, sessionId string) {
	sessionRegistry.Store(userId, sessionId)
}

// GetMockSessionId 读取用户当前关联的 mock sessionId。
// 第二个返回值为 false 表示该用户尚未触发过 Chat()。
func GetMockSessionId(userId string) (string, bool) {
	v, ok := sessionRegistry.Load(userId)
	if !ok {
		return "", false
	}
	return v.(string), true
}

// ResetMockSessions 暴露给测试或需要清空状态的场景。
func ResetMockSessions() {
	sessionRegistry.Range(func(k, _ any) bool {
		sessionRegistry.Delete(k)
		return true
	})
	completedSessions.Range(func(k, _ any) bool {
		completedSessions.Delete(k)
		return true
	})
}

// completedSessions 记录哪些 mock session 的 SSE 流已结束。
// 前端轮询 session 详情时根据此标记决定 Status: 0（已完成）或 1（进行中）。
var completedSessions sync.Map

// MarkStreamCompleted 在 SSE 流结束时调用，标记 session 已完成。
// 由 controller/chat.go 的 Dispatch 在消费完 mock channel 后触发。
func MarkStreamCompleted(sessionId string) {
	completedSessions.Store(sessionId, true)
}

// IsStreamCompleted 查询 mock session 的 SSE 流是否已结束。
func IsStreamCompleted(sessionId string) bool {
	_, ok := completedSessions.Load(sessionId)
	return ok
}

// sessionStatus 根据 SSE 流是否已完成返回 session 的 Status 字段值。
// 0 = 已完成, 1 = 进行中。
func sessionStatus(sessionId string) uint8 {
	if IsStreamCompleted(sessionId) {
		return 0
	}
	return 1
}

// mockSessionItemTitle 是 ListSessions 返回的硬编码标题。
const mockSessionItemTitle = "代码审查 · GoRaven 项目"

// BuildMockSessionList 按“用户是否调过 Chat()”组装会话列表：
//   · 未调过：返回空 list，totalCount=0
//   · 调过：  返回单条 list，sessionId 与 Chat 接口保持一致
func BuildMockSessionList(userId string) *infra.PageResponse {
	sessionId, ok := GetMockSessionId(userId)
	if !ok {
		return &infra.PageResponse{
			List:       []vo.SessionListItem{},
			TotalPage:  0,
			TotalCount: 0,
			Page:       1,
			PageSize:   20,
		}
	}
	now := time.Now()
	return &infra.PageResponse{
		List: []vo.SessionListItem{
			{
				SessionId:    sessionId,
				Title:        mockSessionItemTitle,
				Status:       sessionStatus(sessionId),
				PersonaId:    0,
				Project:      "",
				LastChatTime: now,
				Created:      now,
			},
		},
		TotalPage:  1,
		TotalCount: 1,
		Page:       1,
		PageSize:   20,
	}
}

// BuildMockSessionDetail 按“sessionId 是否属于当前用户”组装详情：
//   · 用户未调过 Chat()  或  sessionId 不匹配：返回 nil（controller 走 404）
//   · 匹配：返回一条 mock 详情
func BuildMockSessionDetail(sessionId, userId string) *vo.SessionDetailRsp {
	stored, ok := GetMockSessionId(userId)
	if !ok || stored != sessionId {
		return nil
	}
	now := time.Now()
	return &vo.SessionDetailRsp{
		SessionId:     sessionId,
		Title:         mockSessionItemTitle,
		Status:        sessionStatus(sessionId),
		PersonaId:     0,
		Project:       "",
		AIModelId:          1,
		ContextTokens:      1234,
		PromptCachedTokens: 0,
		McpIds:        []int{},
		SkillIds:      []int{},
		LastChatTime:  now,
		Created:       now,
		ModelName:     "DeepSeek · deepseek-chat",
		ModelIcon:     "/logos/deepseek.svg",
		ContextLimit:  32768,
	}
}
