package util

import (
	"context"
	"net/http"
	"testing"
)

type captureTransport struct {
	req *http.Request
}

func (c *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.req = req
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

func TestWithConversationHeader_Empty(t *testing.T) {
	ctx := WithConversationHeader(context.Background(), "X-Opencode-Session", "conv-outer")
	// 空 pair 是显式边界：覆盖外层设置，本次运行不注入
	suppressed := WithConversationHeader(ctx, "", "")
	if k, v := ConversationHeaderFromCtx(suppressed); k != "" || v != "" {
		t.Fatalf("expected empty pair after override, got %q/%q", k, v)
	}
	if k, v := ConversationHeaderFromCtx(ctx); k != "X-Opencode-Session" || v != "conv-outer" {
		t.Fatalf("outer ctx should be untouched, got %q/%q", k, v)
	}
}

func TestConversationHeaderFromCtx(t *testing.T) {
	if k, v := ConversationHeaderFromCtx(context.Background()); k != "" || v != "" {
		t.Fatalf("expected empty pair, got %q/%q", k, v)
	}
	ctx := WithConversationHeader(context.Background(), "X-Opencode-Session", "conv-1")
	if k, v := ConversationHeaderFromCtx(ctx); k != "X-Opencode-Session" || v != "conv-1" {
		t.Fatalf("expected X-Opencode-Session/conv-1, got %q/%q", k, v)
	}
	if k, v := ConversationHeaderFromCtx(nil); k != "" || v != "" {
		t.Fatalf("nil ctx should return empty pair, got %q/%q", k, v)
	}
}

func TestGoravenTransport_ConversationHeader(t *testing.T) {
	capture := &captureTransport{}
	transport := &goravenTransport{debug: false, Transport: capture}

	req, _ := http.NewRequestWithContext(WithConversationHeader(context.Background(), "X-Opencode-Session", "sess-42"), http.MethodPost, "https://api.example.com/v1/chat", nil)
	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	if got := capture.req.Header.Get("X-Opencode-Session"); got != "sess-42" {
		t.Fatalf("expected header sess-42, got %q", got)
	}

	req, _ = http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.example.com/v1/chat", nil)
	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	if got := capture.req.Header.Get("X-Opencode-Session"); got != "" {
		t.Fatalf("expected no header, got %q", got)
	}
}
