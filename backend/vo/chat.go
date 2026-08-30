package vo

// ChatReq POST /api/chat 发送消息请求
type ChatReq struct {
	SessionId       *string  `json:"sessionId"`                     // 会话ID，nil=新建会话
	Content         string   `json:"content" validate:"required"`   // 消息内容（必填）
	Attachments     []string `json:"attachments"`                   // 上传文件uploadId列表
	AIModelId       int      `json:"aiModelId"`                     // 模型ID，0表示使用默认模型（后端从默认池随机选取）
	PersonaId       *int     `json:"personaId"`                     // 角色ID
	McpIds          []int    `json:"mcpIds"`                        // MCP ID列表（无角色时使用）
	SkillIds        []int    `json:"skillIds"`                      // 技能ID列表（无角色时使用）
	Reasoning       int      `json:"reasoning"`                     // 思考模式：0无思考 1深度思考
	Project         string   `json:"project"`                       // 项目目录名称，空表示无项目
	TeamProjectId   *int     `json:"sharedProjectId"`               // 团队项目ID，与Project互斥
}

// ChatRsp POST /api/chat 发送消息响应
type ChatRsp struct {
	SessionId string              `json:"sessionId"`          // 会话ID
	Session   *SessionDetailRsp   `json:"session,omitempty"`  // 会话详情（新建/复用会话时返回，前端可直接使用，省去额外 getSessionDetail 请求）
}

// ChatStopReq POST /api/chat/stop 停止生成请求
type ChatStopReq struct {
	SessionId string `json:"sessionId"` // 会话ID（必填）
}

// ChatCompressReq POST /api/chat/compress 手动压缩上下文请求
type ChatCompressReq struct {
	SessionId string `json:"sessionId"` // 会话ID（必填）
}

// ChatCompressRsp POST /api/chat/compress 手动压缩上下文响应
type ChatCompressRsp struct {
	TaskId string `json:"taskId"` // 压缩任务ID，用于轮询状态
}

// CompressTaskStatus 压缩任务状态
const (
	CompressTaskStatusRunning = "running" // 压缩中
	CompressTaskStatusDone    = "done"    // 压缩完成
	CompressTaskStatusFailed  = "failed"  // 压缩失败
)

// ChatCompressPollRsp GET /api/chat/compress/:taskId 压缩任务轮询响应
type ChatCompressPollRsp struct {
	Status  string `json:"status"`  // 任务状态：running/done/failed
	Message string `json:"message"` // 失败时的错误信息
}
