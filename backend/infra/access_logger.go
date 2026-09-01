package infra

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/iris/v12/context"
	"github.com/8treenet/freedom/middleware"
	"github.com/kataras/golog"
)

// requestBodyMaxLen 单条请求体写入日志的最大长度，与 freedom 默认值保持一致。
const requestBodyMaxLen = 512

// maskPlaceholder 敏感字段脱敏后的占位值。
const maskPlaceholder = "***"

// sensitiveFieldKeys 需要脱敏的字段名（小写子串匹配）。
var sensitiveFieldKeys = []string{"password", "passwd", "secret", "credential"}

// NewAccessLogger 返回带敏感信息脱敏的请求日志中间件。
//
// 行为与 middleware.NewRequestLogger 一致：为每个条请求绑定独立的 trace logger，
// 并在请求结束时输出 ACCESS 日志；区别在于写入日志前会对 JSON 请求体中的
// 敏感字段（如 password）脱敏，避免明文密码落入日志。
func NewAccessLogger(traceIDName string) func(context.Context) {
	return func(ctx context.Context) {
		startTime := time.Now()

		var reqBody []byte
		contentType := ctx.Request().Header.Get("Content-Type")
		isJSON := strings.HasPrefix(strings.ToLower(contentType), "application/json")
		if isJSON {
			reqBody, _ = io.ReadAll(ctx.Request().Body)
			ctx.Request().Body.Close()
			ctx.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		work := freedom.ToWorker(ctx)
		freelog := middleware.NewLogger(traceIDName, work.Bus().Get(traceIDName))
		for _, level := range []golog.Level{golog.DebugLevel, golog.WarnLevel, golog.ErrorLevel, golog.FatalLevel} {
			freelog.SetCallerLevel(level)
		}
		work.SetLogger(freelog)

		rawQuery := ctx.Request().URL.Query()
		ctx.Next()

		fieldsMessage := golog.Fields{}
		fieldsMessage["status"] = strconv.Itoa(ctx.GetStatusCode())
		fieldsMessage["latency"] = time.Since(startTime).String()
		fieldsMessage["method"] = ctx.Method()
		fieldsMessage["path"] = ctx.Path()
		if len(rawQuery) > 0 {
			fieldsMessage["query"] = rawQuery.Encode()
		}
		if traceInfo := work.Bus().Get(traceIDName); traceInfo != "" {
			fieldsMessage[traceIDName] = traceInfo
		}
		if len(reqBody) > 0 {
			masked := maskSensitiveJSON(reqBody)
			masked = masked[:min(len(masked), requestBodyMaxLen)]
			if masked != "" {
				fieldsMessage["request"] = masked
			}
		}
		ctx.Application().Logger().Info("[ACCESS]", fieldsMessage)
	}
}

// maskSensitiveJSON 对 JSON 请求体中的敏感字段值做脱敏。
// 非法 JSON 返回占位文本，避免意外泄露；空内容返回空字符串。
func maskSensitiveJSON(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "[unparseable json body omitted]"
	}
	maskJSONValue(parsed)
	out, err := json.Marshal(parsed)
	if err != nil {
		return "[unparseable json body omitted]"
	}
	return string(out)
}

// maskJSONValue 递归遍历已解析的 JSON 值，将命中敏感字段名的值替换为占位符。
func maskJSONValue(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			if isSensitiveField(k) {
				val[k] = maskPlaceholder
				continue
			}
			maskJSONValue(child)
		}
	case []interface{}:
		for _, child := range val {
			maskJSONValue(child)
		}
	}
}

// isSensitiveField 判断字段名是否命中敏感字段（大小写不敏感的子串匹配）。
func isSensitiveField(key string) bool {
	lower := strings.ToLower(key)
	for _, sensitive := range sensitiveFieldKeys {
		if strings.Contains(lower, sensitive) {
			return true
		}
	}
	return false
}
