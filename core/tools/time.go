package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"goraven/config"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GetCurrentTimeRequest struct {
	Timezone string `json:"timezone,omitempty" jsonschema:"description=IANA timezone name such as Asia/Shanghai or UTC, optional, defaults to the server local timezone"`
}

type GetCurrentTimeResponse struct {
	Datetime string `json:"datetime" jsonschema:"description=Current date and time in RFC3339 format, relay this to the user"`
	Unix     int64  `json:"unix" jsonschema:"description=Unix timestamp in seconds"`
	Timezone string `json:"timezone" jsonschema:"description=Timezone name used for the result"`
	Weekday  string `json:"weekday" jsonschema:"description=Day of the week in English, e.g. Monday"`
	IsDst    bool   `json:"is_dst" jsonschema:"description=Whether daylight saving time is currently in effect"`
}

type GetCurrentTime struct {
	Name string
	Desc string
}

const (
	GetCurrentTimeToolDesc = `Get the current date and time. Optionally pass an IANA timezone name; if omitted, the local timezone is used. Always call this tool when you need to know the current time instead of guessing.`

	GetCurrentTimeToolDescChinese = `获取当前日期和时间。可选传入IANA时区名称；不传时使用本地时区。需要知道当前时间时必须调用此工具，禁止凭空猜测。`
)

func NewGetCurrentTime() (tool.InvokableTool, error) {
	desc := GetCurrentTimeToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = GetCurrentTimeToolDescChinese
	}

	t := &GetCurrentTime{
		Name: "goraven_get_current_time",
		Desc: desc,
	}

	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

func (g *GetCurrentTime) Invoke(ctx context.Context, req *GetCurrentTimeRequest) (*GetCurrentTimeResponse, error) {
	loc := time.Local
	if tz := strings.TrimSpace(req.Timezone); tz != "" && tz != "local" {
		var err error
		loc, err = time.LoadLocation(tz)
		if err != nil {
			return nil, automationToolErr(
				"invalid timezone %q, expected an IANA timezone name such as Asia/Shanghai or UTC",
				"无效的时区 %q，应为 IANA 时区名称，如 Asia/Shanghai 或 UTC", req.Timezone)
		}
	}

	now := time.Now().In(loc)
	return &GetCurrentTimeResponse{
		Datetime: now.Format(time.RFC3339),
		Unix:     now.Unix(),
		Timezone: loc.String(),
		Weekday:  now.Weekday().String(),
		IsDst:    now.IsDST(),
	}, nil
}
