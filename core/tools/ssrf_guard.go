package tools

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// validatePublicHTTPURL 校验外部抓取 URL，防止 SSRF：
//   - 仅允许 http/https 协议（阻断 file://、gopher:// 等）
//   - 目标地址（IP 直连或域名解析结果）不得为回环、私网、链路本地等内网地址
//
// 注意：域名场景下在校验与实际请求之间存在 DNS 重绑定（rebinding）的 TOCTOU 窗口，
// 按当前架构已能拦截绝大部分误用；彻底解决需要在传输层固定解析后的 IP。
func validatePublicHTTPURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("scheme %q is not allowed, only http and https are supported", u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url has no host")
	}

	// host 为 IP 字面量，直接校验
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return fmt.Errorf("access to private or reserved address is not allowed: %s", host)
		}
		return nil
	}

	// 域名场景：解析后对所有结果逐一校验
	addrs, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %q: %w", host, err)
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if !isPublicIP(ip) {
			return fmt.Errorf("host %q resolves to private or reserved address: %s", host, addr)
		}
	}
	return nil
}

// isPublicIP 判断 IP 是否为可对外访问的公网地址。
func isPublicIP(ip net.IP) bool {
	return !(ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified())
}
