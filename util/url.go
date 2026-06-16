package util

import "strings"

// EncodeQuery URL 编码查询参数
func EncodeQuery(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(s, " ", "%20"),
			"/", "%2F",
		),
		"?", "%3F",
	)
}

// EncodePath URL 编码路径段
func EncodePath(s string) string {
	return strings.ReplaceAll(s, "/", "%2F")
}
