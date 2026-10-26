package mock

import (
	"testing"

	"goraven/backend/vo"
)

// TestSessionListEmpty 校验未调过 Chat() 时会话列表返回空。
func TestSessionListEmpty(t *testing.T) {
	ResetMockSessions()

	const user = "test-user-empty"
	rsp := BuildMockSessionList(user)
	if rsp == nil {
		t.Fatal("expected non-nil page response")
	}
	list, ok := rsp.List.([]vo.SessionListItem)
	if !ok {
		t.Fatalf("expected []vo.SessionListItem, got %T", rsp.List)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list before Chat() is called, got %d items", len(list))
	}
	if rsp.TotalCount != 0 || rsp.TotalPage != 0 {
		t.Errorf("expected totalCount/totalPage = 0, got %d/%d", rsp.TotalCount, rsp.TotalPage)
	}
}

// TestSessionListAfterChat 校验调过 Chat() 后会话列表返回单条且 sessionId 一致。
func TestSessionListAfterChat(t *testing.T) {
	ResetMockSessions()

	const user = "test-user-after-chat"
	MarkSessionCreated(user, ChatMockSessionId)

	rsp := BuildMockSessionList(user)
	list, ok := rsp.List.([]vo.SessionListItem)
	if !ok {
		t.Fatalf("expected []vo.SessionListItem, got %T", rsp.List)
	}
	if len(list) != 1 {
		t.Fatalf("expected exactly 1 item, got %d", len(list))
	}
	if list[0].SessionId != ChatMockSessionId {
		t.Errorf("sessionId mismatch: got %q want %q", list[0].SessionId, ChatMockSessionId)
	}
	if list[0].Status != 1 {
		t.Errorf("expected status=1 (in progress), got %d", list[0].Status)
	}
	if rsp.TotalCount != 1 || rsp.TotalPage != 1 {
		t.Errorf("expected totalCount/totalPage = 1, got %d/%d", rsp.TotalCount, rsp.TotalPage)
	}
}

// TestSessionListIsolation 校验不同用户的 mock 会话互不干扰。
func TestSessionListIsolation(t *testing.T) {
	ResetMockSessions()

	MarkSessionCreated("user-A", "session-A")
	if _, ok := GetMockSessionId("user-B"); ok {
		t.Error("user-B should not see user-A's session")
	}
	got, ok := GetMockSessionId("user-A")
	if !ok || got != "session-A" {
		t.Errorf("user-A should see session-A, got %q ok=%v", got, ok)
	}
}

// TestSessionDetailNotFound 校验 sessionId 不匹配时返回 nil，模拟 404。
func TestSessionDetailNotFound(t *testing.T) {
	ResetMockSessions()

	// 用户没调过 Chat()
	if d := BuildMockSessionDetail("any", "user-x"); d != nil {
		t.Errorf("expected nil for user without chat, got %+v", d)
	}

	// 用户调过 Chat()，但 sessionId 不一致
	MarkSessionCreated("user-y", "real-session")
	if d := BuildMockSessionDetail("other-session", "user-y"); d != nil {
		t.Errorf("expected nil for mismatched sessionId, got %+v", d)
	}
}

// TestSessionDetailOK 校验 sessionId 匹配时返回非空详情。
func TestSessionDetailOK(t *testing.T) {
	ResetMockSessions()

	const user = "user-detail"
	MarkSessionCreated(user, ChatMockSessionId)

	d := BuildMockSessionDetail(ChatMockSessionId, user)
	if d == nil {
		t.Fatal("expected non-nil detail")
	}
	if d.SessionId != ChatMockSessionId {
		t.Errorf("sessionId mismatch: %q", d.SessionId)
	}
	if d.Status != 1 {
		t.Errorf("expected status=1, got %d", d.Status)
	}
	if d.ModelName == "" {
		t.Error("expected non-empty modelName")
	}
}
