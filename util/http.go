package util

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/8treenet/freedom"
)

var (
	goravenHTTPClient = &http.Client{
		Transport: &goravenTransport{
			debug:     false,
			Transport: http.DefaultTransport,
		},
	}

	proxyClients      map[string]*http.Client = map[string]*http.Client{}
	proxyClientsMutex sync.Mutex
)

type convHeader struct {
	key, value string
}

type convHeaderKey struct{}

// WithConversationHeader 设置供应商请求的会话归并 header：key 为 header 名（如 X-Opencode-Session），value 为逻辑运行 ID。
// 包装即生效边界：覆盖外层设置；key 或 value 为空表示本次运行不注入。
func WithConversationHeader(ctx context.Context, key, value string) context.Context {
	return context.WithValue(ctx, convHeaderKey{}, convHeader{key: key, value: value})
}

// ConversationHeaderFromCtx 读取会话归并 header 的 key/value，未设置时返回空串
func ConversationHeaderFromCtx(ctx context.Context) (key, value string) {
	if ctx == nil {
		return "", ""
	}
	if h, ok := ctx.Value(convHeaderKey{}).(convHeader); ok {
		return h.key, h.value
	}
	return "", ""
}

func GetGoRavenHTTPClient() *http.Client {
	return goravenHTTPClient
}

func NewDebugHTTPClient() *http.Client {
	return &http.Client{
		Transport: &goravenTransport{
			debug:     true,
			Transport: http.DefaultTransport,
		},
	}
}

// GetProxyHTTPClient
// proxyURL: http://127.0.0.1:33210
func GetProxyHTTPClient(proxyURL string) *http.Client {
	proxyClientsMutex.Lock()
	defer proxyClientsMutex.Unlock()

	if c, ok := proxyClients[proxyURL]; ok {
		return c
	}

	proxyurl, _ := url.Parse(proxyURL)
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyurl),
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	ts := &goravenTransport{
		debug:     false,
		Transport: transport,
	}
	client := &http.Client{Transport: ts}
	proxyClients[proxyURL] = client
	return client
}

type goravenTransport struct {
	debug     bool
	Transport http.RoundTripper
}

func (t *goravenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("HTTP-Referer", "https://goraven.dev")
	req.Header.Set("X-Title", "GoRaven")
	req.Header.Set("X-OpenRouter-Title", "GoRaven")
	// 供应商会话归并 header（如 X-Opencode-Session），header 名由 AgentParam 配置，空则不注入
	if key, value := ConversationHeaderFromCtx(req.Context()); key != "" && value != "" {
		req.Header.Set(key, value)
	}
	if !t.debug {
		return t.Transport.RoundTrip(req)
	}

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		freedom.Logger().Errorf("[HTTP Debug] DumpRequestOut error: %v", err)
	} else {
		freedom.Logger().Debug("[HTTP Debug] %s", string(dump))
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		freedom.Logger().Errorf("[HTTP Debug] RoundTrip error: %v", err)
		return resp, err
	}

	dumpResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		freedom.Logger().Errorf("[HTTP Debug] DumpResponse error: %v", err)
	} else {
		freedom.Logger().Debug("[HTTP Debug] %s", string(dumpResp))
	}

	return resp, nil
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}
