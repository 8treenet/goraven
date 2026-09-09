package agent

import (
	"encoding/json"
	"goraven/backend/po"
	"goraven/core/tools"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// MediaItem 多模态媒体附件条目（直接模式），与 po.Message.Media JSON 对应。
// URL 为 FileLink 签发的 72h 外链，重建历史时直接复用，不重新签发。
type MediaItem struct {
	Type string `json:"type"` // image/video/audio（对应 tools.MediaType*）
	Path string `json:"path"` // 沙盒相对路径，如 /temp/xxx.jpg
	URL  string `json:"url"`  // 可公网访问的文件外链
}

// mediaPart 将 MediaItem 转为 schema 输入部件（URL 直发，不用 base64）
func mediaPart(m MediaItem) (schema.MessageInputPart, bool) {
	url := m.URL
	common := schema.MessagePartCommon{URL: &url}
	switch m.Type {
	case tools.MediaTypeImage:
		return schema.MessageInputPart{
			Type:  schema.ChatMessagePartTypeImageURL,
			Image: &schema.MessageInputImage{MessagePartCommon: common},
		}, true
	case tools.MediaTypeVideo:
		return schema.MessageInputPart{
			Type:  schema.ChatMessagePartTypeVideoURL,
			Video: &schema.MessageInputVideo{MessagePartCommon: common},
		}, true
	case tools.MediaTypeAudio:
		return schema.MessageInputPart{
			Type:  schema.ChatMessagePartTypeAudioURL,
			Audio: &schema.MessageInputAudio{MessagePartCommon: common},
		}, true
	}
	return schema.MessageInputPart{}, false
}

// BuildUserMessage 构建用户消息：无媒体时纯文本；有媒体时 text 部件 + URL 媒体部件
func BuildUserMessage(content string, media []MediaItem) adk.Message {
	if len(media) == 0 {
		return schema.UserMessage(content)
	}
	parts := make([]schema.MessageInputPart, 0, len(media)+1)
	parts = append(parts, schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeText,
		Text: content,
	})
	for _, m := range media {
		if p, ok := mediaPart(m); ok {
			parts = append(parts, p)
		}
	}
	return &schema.Message{Role: schema.User, UserInputMultiContent: parts}
}

func BuildHistoryFromMessages(list []*po.Message) (result []adk.Message) {
	for _, v := range list {
		var msg adk.Message
		switch v.RoleType {
		case po.RoleTypeUser:
			if v.Media != "" {
				var media []MediaItem
				if json.Unmarshal([]byte(v.Media), &media) == nil && len(media) > 0 {
					msg = BuildUserMessage(v.Content, media)
					break
				}
			}
			msg = schema.UserMessage(v.Content)
		case po.RoleTypeAssistant:
			var toolCalls []schema.ToolCall
			if v.Ext != "" {
				var ext po.AssistantExt
				if json.Unmarshal([]byte(v.Ext), &ext) == nil {
					for _, tc := range ext.ToolCalls {
						toolCalls = append(toolCalls, schema.ToolCall{
							ID:   tc.ID,
							Type: "function",
							Function: schema.FunctionCall{
								Name:      tc.Name,
								Arguments: tc.Arguments,
							},
						})
					}
				}
			}
			msg = schema.AssistantMessage(v.Content, toolCalls)
			msg.ReasoningContent = v.ReasoningContent
		case po.RoleTypeTool:
			var toolCallID, toolName string
			if v.Ext != "" {
				var ext po.ToolExt
				if json.Unmarshal([]byte(v.Ext), &ext) == nil {
					toolCallID = ext.ToolCallID
					toolName = ext.ToolName
				}
			}
			msg = schema.ToolMessage(v.Content, toolCallID, schema.WithToolName(toolName))
			msg.ReasoningContent = v.ReasoningContent
		case po.RoleTypeSummary:
			msg = schema.UserMessage(v.Content)
		}
		msg.Extra = map[string]any{
			"timestamp": v.Timestamp,
			"msgId":     v.MsgId,
			"roundId":   v.RoundId,
		}
		if v.RoleType == po.RoleTypeSummary {
			msg.Extra["isSummary"] = true
		}

		result = append(result, msg)
	}
	return
}
