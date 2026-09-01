package vo

import (
	"time"
)

// TeamProjectInfo 会话中关联的团队项目信息
type TeamProjectInfo struct {
	Id          int    `json:"id"`          // 团队项目ID
	CreatorId   string `json:"creatorId"`   // 创建者 userId
	CreatorName string `json:"creatorName"` // 创建者昵称
	ProjectName string `json:"projectName"` // 项目目录名
	Description string `json:"description"` // 项目简介
}

// SessionUpdateReq 会话更新请求（只传需修改字段，nil 字段不修改）
type SessionUpdateReq struct {
	Title      *string `json:"title"`      // 新标题，nil 不修改
	IsArchived *int    `json:"isArchived"` // 归档状态：0否 1是，nil 不修改
}

// SessionListReq 会话列表请求
type SessionListReq struct {
	Page            int    `url:"page"`            // 页码，默认 1
	PageSize        int    `url:"pageSize"`        // 每页条数，默认 20
	PersonaId       *int   `url:"personaId"`       // 按角色筛选，nil 不筛选
	Project         string `url:"project"`         // 按个人项目目录名筛选，空不筛选
	SharedProjectId *int   `url:"sharedProjectId"` // 按团队项目ID筛选，nil 不筛选
}

// SessionListItem 会话列表条目（侧边栏"所有对话"使用）
type SessionListItem struct {
	SessionId     string             `json:"sessionId"`              // 会话ID
	Title         string             `json:"title"`                  // 会话标题
	Status        uint8              `json:"status"`                 // 会话状态：0正常 1进行中
	PersonaId     int                `json:"personaId"`              // 角色ID，0表示无角色绑定
	Project       string             `json:"project"`                // 项目目录名称，空表示无项目
	TeamProject   *TeamProjectInfo   `json:"sharedProject,omitempty"` // 团队项目，team_project_id>0 时非nil
	LastChatTime  time.Time          `json:"lastChatTime"`           // 最后聊天时间
	Created       time.Time          `json:"created"`                // 创建时间
}

// SessionDetailRsp 会话详情响应（含模型、MCP/技能快照）
type SessionDetailRsp struct {
	SessionId             string             `json:"sessionId"`                      // 会话ID
	Title                 string             `json:"title"`                          // 会话标题
	Status                uint8              `json:"status"`                         // 会话状态：0正常 1进行中
	PersonaId             int                `json:"personaId"`                      // 角色ID，0表示无角色绑定
	Project               string             `json:"project"`                        // 项目目录名称，空表示无项目
	TeamProject           *TeamProjectInfo   `json:"sharedProject,omitempty"`         // 团队项目，team_project_id>0 时非nil
	AIModelId             int                `json:"aiModelId"`                      // 使用模型的ID
	ContextTokens         int                `json:"contextTokens"`                  // 当前上下文长度
	PromptTokensCount     int                `json:"promptTokensCount" gorm:"column:prompt_tokens_count"`     //累计promptTokensCount
	CompletionTokensCount int                `json:"completionTokensCount" gorm:"column:completion_tokens_count"` //累计completionTokensCount
	PromptCachedTokens    int                `json:"promptCachedTokens"`             //累计缓存promptTokens
	McpIds                []int              `json:"mcpIds"`                         // MCP配置ID列表，有角色时为空
	SkillIds              []int              `json:"skillIds"`                       // 技能ID列表，有角色时为空
	LastChatTime          time.Time          `json:"lastChatTime"`                   // 最后聊天时间
	Created               time.Time          `json:"created"`                        // 创建时间

	// 关联信息（service 层组装）
	ModelName    string `json:"modelName,omitempty"`    // 模型名称（providerDisplayName - modelName）
	ModelIcon    string `json:"modelIcon,omitempty"`    // 模型图标（与 modelName 同一来源，避免前端从可见模型列表回查不到）
	PersonaName  string `json:"personaName,omitempty"`  // 角色名称
	PersonaIcon  string `json:"personaIcon,omitempty"`  // 角色图标
	ContextLimit int    `json:"contextLimit,omitempty"` // 模型上下文长度上限（tokens）
}

// MessageItem 消息条目
type MessageItem struct {
	MsgId            string                 `json:"msgId"`            // 消息ID
	RoundId          string                 `json:"roundId"`          // 轮次ID
	ContextState     uint8                  `json:"contextState"`     // 上下文状态：0=发送给LLM 1=跳过(仅展示)
	Content          string                 `json:"content"`          // 消息内容
	ReasoningContent []MessageReasoningItem `json:"reasoningContent"` // 推理/思考内容，开启思考且 assistant 类型才有
	RoleType         string                 `json:"roleType"`         // 角色：user/assistant
	Created          string                 `json:"created"`          // 创建时间
}

// EventTool 工具事件的展示信息
type MessageReasoningToolCallsInfo struct {
	Name        string `json:"name"`        // 工具名称
	DisplayName string `json:"displayName"` // 本地化展示名称
	Icon        string `json:"icon"`        // emoji 图标
	Action      string `json:"action"`      // 本地化动作描述
}

type MessageReasoningItem struct {
	EventType string                         `json:"eventType"`
	Content   string                         `json:"content,omitempty"`
	Tool      *MessageReasoningToolCallsInfo `json:"tool,omitempty"`
}
