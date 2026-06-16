package agent

const (
	SSEEventTypeReasoning = "reasoning"
	SSEEventTypeTool      = "tool"
	SSEEventTypeEnd       = "end"
)

type SSEEvent struct {
	Type string `json:"type"`
}
