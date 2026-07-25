package mock

import (
	"sync"
	"time"

	"goraven/backend/infra"
	"goraven/backend/vo"
)

const ChatMockSessionId = "mock-session-2026-demo"

var sessionRegistry sync.Map

func MarkSessionCreated(userId, sessionId string) {
	sessionRegistry.Store(userId, sessionId)
}

func GetMockSessionId(userId string) (string, bool) {
	v, ok := sessionRegistry.Load(userId)
	if !ok {
		return "", false
	}
	return v.(string), true
}

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

var completedSessions sync.Map

func MarkStreamCompleted(sessionId string) {
	completedSessions.Store(sessionId, true)
}

func IsStreamCompleted(sessionId string) bool {
	_, ok := completedSessions.Load(sessionId)
	return ok
}

func sessionStatus(sessionId string) uint8 {
	if IsStreamCompleted(sessionId) {
		return 0
	}
	return 1
}

const mockSessionItemTitle = "代码审查 · GoRaven 项目"

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
		AIModelId:     1,
		ContextTokens: 1234,
		McpIds:        []int{},
		SkillIds:      []int{},
		LastChatTime:  now,
		Created:       now,
		ModelName:     "DeepSeek · deepseek-chat",
		ContextLimit:  32768,
	}
}
