package vo

type ChatReq struct {
	SessionId       *string  `json:"sessionId"`
	Content         string   `json:"content" validate:"required"`
	Attachments     []string `json:"attachments"`
	AIModelId       int      `json:"aiModelId" validate:"required"`
	PersonaId       *int     `json:"personaId"`
	McpIds          []int    `json:"mcpIds"`
	SkillIds        []int    `json:"skillIds"`
	Reasoning       int      `json:"reasoning"`
	Project         string   `json:"project"`
	SharedProjectId *int     `json:"sharedProjectId"`
}

type ChatRsp struct {
	SessionId string            `json:"sessionId"`
	Session   *SessionDetailRsp `json:"session,omitempty"`
}

type ChatStopReq struct {
	SessionId string `json:"sessionId"`
}

type ChatCompressReq struct {
	SessionId string `json:"sessionId"`
}

type ChatCompressRsp struct {
	TaskId string `json:"taskId"`
}

const (
	CompressTaskStatusRunning = "running"
	CompressTaskStatusDone    = "done"
	CompressTaskStatusFailed  = "failed"
)

type ChatCompressPollRsp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
