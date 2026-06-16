package util

import (
	"encoding/json"
	"fmt"
)

// MapToJSON 将 map[string]string 转为 JSON 字符串
func MapToJSON(m map[string]string) (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MapToEnvJSON 将 map[string]string 转为环境变量 JSON 数组字符串
// 如 {"GITHUB_TOKEN":"xxx"} → ["GITHUB_TOKEN=xxx"]
func MapToEnvJSON(m map[string]string) (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	slice := MapToEnvSlice(m)
	data, err := json.Marshal(slice)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MapToEnvSlice 将 map[string]string 转为环境变量数组
// 如 {"GITHUB_TOKEN":"xxx"} → ["GITHUB_TOKEN=xxx"]
func MapToEnvSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// SliceToJSON 将 []string 转为 JSON 数组字符串
func SliceToJSON(s []string) (string, error) {
	if len(s) == 0 {
		return "", nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MaskJSONValues 对 JSON 字符串中的值进行脱敏
// 支持 JSON 对象和 JSON 数组格式
// 对象: {"Authorization":"Bearer sk-xxx"} → {"Authorization":"Be****"}
// 数组: ["KEY=secret"] → ["KE****"]
// 如果不是有效 JSON 或为空则返回 "****"
func MaskJSONValues(raw string) string {
	if raw == "" {
		return ""
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err == nil {
		for k, v := range obj {
			if str, ok := v.(string); ok {
				obj[k] = maskStr(str)
			}
		}
		result, err := json.Marshal(obj)
		if err == nil {
			return string(result)
		}
	}

	var arr []interface{}
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		for i, v := range arr {
			if str, ok := v.(string); ok {
				arr[i] = maskStr(str)
			}
		}
		result, err := json.Marshal(arr)
		if err == nil {
			return string(result)
		}
	}

	return "****"
}

// maskStr 对字符串脱敏，保留前2字符
func maskStr(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****"
}

// TruncateRunes 按 Unicode 字符数截断字符串，超出 maxLen 时追加 "..."
func TruncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// PtrFloat64 辅助函数，取 float64 指针
func PtrFloat64(v float64) *float64 {
	return &v
}

// DedupStrings 字符串数组去重，保持首次出现的顺序
func DedupStrings(ss []string) []string {
	if len(ss) <= 1 {
		return ss
	}
	seen := make(map[string]struct{}, len(ss))
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

// IntFromIFace 从 interface{} 中提取 int 值，兼容 int/uint8/int8/int64/uint64/uint 等整数类型。
// 避免 map[string]interface{} 中 uint8(1) == int(1) 因类型不匹配返回 false 的问题。
func IntFromIFace(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case uint8:
		return int(n)
	case int8:
		return int(n)
	case int64:
		return int(n)
	case uint64:
		return int(n)
	case uint:
		return int(n)
	default:
		return 0
	}
}
