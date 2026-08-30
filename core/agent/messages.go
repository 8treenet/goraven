package agent

import (
	"encoding/json"
	"goraven/backend/po"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func BuildHistoryFromMessages(list []*po.Message) (result []adk.Message) {
	for _, v := range list {
		var msg adk.Message
		switch v.RoleType {
		case po.RoleTypeUser:
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
