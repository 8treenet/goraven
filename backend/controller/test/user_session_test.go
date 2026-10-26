package test

import (
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// TestListSessions 获取会话列表 GET /api/sessions
func TestListSessions(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestListSessionsWithPersonaFilter 按角色筛选会话列表 GET /api/sessions?personaId=1
func TestListSessionsWithPersonaFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions?personaId=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetSession 获取会话详情 GET /api/sessions/:sessionId
func TestGetSession(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/ddd123").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetSessionMessages 获取会话历史消息 GET /api/sessions/:sessionId/messages
func TestGetSessionMessages(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/784bf78750994cddba767f950ba0596a/messages").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestDeleteSession 删除会话（软删除） DELETE /api/sessions/:sessionId
func TestDeleteSession(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/ddd123").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetSessionNotFound 查询不存在的会话
func TestGetSessionNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/nonexistent-session-id").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
