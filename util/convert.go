package util

import (
	"encoding/json"
	"fmt"
)

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

func MapToEnvSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

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

// 对象: {"Authorization":"Bearer sk-xxx"} → {"Authorization":"Be****"}
// 数组: ["KEY=secret"] → ["KE****"]

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

func maskStr(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****"
}

func TruncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func PtrFloat64(v float64) *float64 {
	return &v
}

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

func DedupInts(ns []int) []int {
	if len(ns) <= 1 {
		return ns
	}
	seen := make(map[int]struct{}, len(ns))
	result := make([]int, 0, len(ns))
	for _, n := range ns {
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			result = append(result, n)
		}
	}
	return result
}

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
