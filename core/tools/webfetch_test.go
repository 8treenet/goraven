package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 确认工具入口确实走安全客户端：默认配置下不能抓取本机 / 内网地址。
func TestWebFetchRejectsInternalURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>internal-only</body></html>"))
	}))
	defer srv.Close()

	w := &WebFetch{}
	if _, err := w.Invoke(context.Background(), &WebFetchRequest{URL: srv.URL}); err == nil {
		t.Fatal("WebFetch fetched a loopback URL; SSRF protection is not effective at tool entry")
	}

	for _, raw := range []string{
		"file:///etc/passwd",
		"http://169.254.169.254/latest/meta-data/",
		"http://127.0.0.1:6379/",
	} {
		if _, err := w.Invoke(context.Background(), &WebFetchRequest{URL: raw}); err == nil {
			t.Errorf("WebFetch accepted unsafe URL %q", raw)
		}
	}
}
