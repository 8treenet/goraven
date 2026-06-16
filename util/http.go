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
	ravenHTTPClient = &http.Client{
		Transport: &ravenTransport{
			debug:     false,
			Transport: http.DefaultTransport,
		},
	}

	proxyClients      map[string]*http.Client = map[string]*http.Client{}
	proxyClientsMutex sync.Mutex
)

func GetRavenHTTPClient() *http.Client {
	return ravenHTTPClient
}

func NewDebugHTTPClient() *http.Client {
	return &http.Client{
		Transport: &ravenTransport{
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
	ts := &ravenTransport{
		debug:     false,
		Transport: transport,
	}
	client := &http.Client{Transport: ts}
	proxyClients[proxyURL] = client
	return client
}

type ravenTransport struct {
	debug     bool
	Transport http.RoundTripper
}

func (t *ravenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("HTTP-Referer", "https://goraven.dev")
	req.Header.Set("X-Title", "Raven")
	req.Header.Set("X-OpenRouter-Title", "Raven")
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
