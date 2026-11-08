package util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidatePublicURLRejectsInternalTargets(t *testing.T) {
	cases := []string{
		"",
		"file:///etc/passwd",
		"gopher://127.0.0.1:6379/_INFO",
		"ftp://example.com/x",
		"http://127.0.0.1:8000/api/ping",
		"http://169.254.169.254/latest/meta-data/iam/security-credentials/",
		"http://10.0.0.1/",
		"http://172.16.0.1/",
		"http://192.168.1.1/",
		"http://0.0.0.0/",
		"http://100.64.0.1/",
		"http://[::1]/",
		"http://[::ffff:127.0.0.1]/",
		"http://user:pass@example.com/",
	}
	for _, raw := range cases {
		if _, err := ValidatePublicURL(raw); err == nil {
			t.Errorf("ValidatePublicURL(%q) should be rejected", raw)
		}
	}
}

func TestCheckPublicIPAcceptsPublicAndRejectsReserved(t *testing.T) {
	if err := CheckPublicIP([]byte{8, 8, 8, 8}); err != nil {
		t.Fatalf("public IP rejected: %v", err)
	}
	if err := CheckPublicIP([]byte{127, 0, 0, 1}); err == nil {
		t.Fatal("loopback IP accepted")
	}
}

// 端到端：默认配置下，即使本机起了 HTTP 服务，WebFetch 也连不上；
// 显式打开 allowPrivateFetch 后才可访问。
func TestSafeHTTPClientBlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("secret-internal-data"))
	}))
	defer srv.Close()

	blocked := NewSafeHTTPClient(3*time.Second, false)
	if _, err := FetchURL(context.Background(), srv.URL, blocked, false); err == nil {
		t.Fatal("SSRF protection bypassed: loopback target was fetched")
	}

	allowed := NewSafeHTTPClient(3*time.Second, true)
	body, err := FetchURL(context.Background(), srv.URL, allowed, true)
	if err != nil {
		t.Fatalf("allowPrivateFetch=true should permit loopback: %v", err)
	}
	if body != "secret-internal-data" {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestFetchURLRejectsOversizedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", MaxFetchResponseBytes+1)))
	}))
	defer srv.Close()

	client := NewSafeHTTPClient(5*time.Second, true)
	if _, err := FetchURL(context.Background(), srv.URL, client, true); err == nil {
		t.Fatal("oversized response was not rejected")
	}
}
