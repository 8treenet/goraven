package util

import (
	"net"
	"net/url"
	"strings"
)

func EncodeQuery(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(s, " ", "%20"),
			"/", "%2F",
		),
		"?", "%3F",
	)
}

func EncodePath(s string) string {
	return strings.ReplaceAll(s, "/", "%2F")
}

func IsLocalOrIPHost(host string) bool {
	h := host
	if i := strings.LastIndex(h, ":"); i >= 0 {
		h = h[:i]
	}
	h = strings.TrimSpace(h)
	if h == "localhost" || strings.HasSuffix(h, ".localhost") {
		return true
	}
	return net.ParseIP(h) != nil
}

func IsLocalOrIPURL(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	return IsLocalOrIPHost(parsed.Host)
}
