package util

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 工具侧拉取网页时的响应体上限。返回内容会进入 LLM 上下文，必须限制大小，
// 否则一个内网大文件或无限流就能把内存与 token 预算打满。
const MaxFetchResponseBytes = 2 << 20 // 2 MiB

const maxFetchRedirects = 3

var blockedFetchCIDRs = func() []*net.IPNet {
	blocks := []string{
		"0.0.0.0/8",       // 本机 / 未指定
		"100.64.0.0/10",   // 运营商级 NAT
		"169.254.0.0/16",  // 链路本地，含云元数据 169.254.169.254
		"192.0.0.0/24",    // IETF 协议保留
		"192.0.2.0/24",    // TEST-NET-1
		"198.18.0.0/15",   // 基准测试
		"198.51.100.0/24", // TEST-NET-2
		"203.0.113.0/24",  // TEST-NET-3
		"224.0.0.0/4",     // 组播
		"240.0.0.0/4",     // 保留
		"::/128",          // IPv6 未指定
		"::1/128",         // IPv6 环回
		"fc00::/7",        // IPv6 唯一本地
		"fe80::/10",       // IPv6 链路本地
		"ff00::/8",        // IPv6 组播
	}
	out := make([]*net.IPNet, 0, len(blocks))
	for _, b := range blocks {
		_, cidr, err := net.ParseCIDR(b)
		if err == nil {
			out = append(out, cidr)
		}
	}
	return out
}()

// CheckPublicIP 判定解析出的 IP 是否允许作为外网抓取目标。
// 覆盖环回、私网、链路本地、组播、未指定与常见保留段。
func CheckPublicIP(ip net.IP) error {
	if ip == nil {
		return fmt.Errorf("empty IP address")
	}
	// IPv4-mapped IPv6 统一按 IPv4 判定，避免 ::ffff:127.0.0.1 绕过。
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return fmt.Errorf("target address %s is not a public IP", ip.String())
	}
	for _, cidr := range blockedFetchCIDRs {
		if cidr.Contains(ip) {
			return fmt.Errorf("target address %s is in reserved range %s", ip.String(), cidr.String())
		}
	}
	return nil
}

// ValidatePublicURL 校验一个待抓取 URL：
//   - scheme 只允许 http / https（挡掉 file:, gopher:, dict: 等）；
//   - 不允许 URL 内嵌凭据；
//   - 主机是 IP 字面量时直接判定，是域名时解析全部 A/AAAA 记录逐一判定。
//
// 注意连接层还有 NewSafeHTTPClient 的二次校验，这里不是唯一防线（防 DNS rebinding）。
func ValidatePublicURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme %q, only http/https are allowed", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("URL has no host")
	}
	if u.User != nil {
		return nil, fmt.Errorf("credentials embedded in URL are not allowed")
	}

	if ip := net.ParseIP(host); ip != nil {
		if err := CheckPublicIP(ip); err != nil {
			return nil, err
		}
		return u, nil
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, fmt.Errorf("resolve host %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("host %q has no addresses", host)
	}
	for _, addr := range addrs {
		if err := CheckPublicIP(addr.IP); err != nil {
			return nil, err
		}
	}
	return u, nil
}

// NewSafeHTTPClient 构造用于「抓取不可信 URL」的 HTTP 客户端。
//   - DialContext 在真正建立连接前再次解析并校验 IP，且固定连到已校验的 IP，
//     避免校验与实际连接之间的 DNS rebinding 窗口；
//   - 重定向逐跳校验（最多 3 跳），跳转目标同样受 scheme / IP 规则约束；
//   - 整体超时由调用方给定。
//
// allowPrivate=true 用于完全离线的内网部署，由管理员在配置里显式开启。
func NewSafeHTTPClient(timeout time.Duration, allowPrivate bool) *http.Client {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("host %q has no addresses", host)
		}
		if !allowPrivate {
			for _, a := range addrs {
				if err := CheckPublicIP(a.IP); err != nil {
					return nil, err
				}
			}
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(addrs[0].IP.String(), port))
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxFetchRedirects {
				return fmt.Errorf("too many redirects")
			}
			if !allowPrivate {
				if _, err := ValidatePublicURL(req.URL.String()); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// FetchURL 用安全客户端抓取网页文本。调用方应先确认 URL 已通过 ValidatePublicURL。
func FetchURL(ctx context.Context, raw string, client *http.Client, allowPrivate bool) (string, error) {
	if client == nil {
		return "", fmt.Errorf("nil http client")
	}
	if !allowPrivate {
		if _, err := ValidatePublicURL(raw); err != nil {
			return "", err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "GoRaven-WebFetch/1.0")
	req.Header.Set("Accept", "text/html,text/plain,application/json;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxFetchResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if len(body) > MaxFetchResponseBytes {
		return "", fmt.Errorf("response exceeds %d bytes", MaxFetchResponseBytes)
	}
	return string(body), nil
}
