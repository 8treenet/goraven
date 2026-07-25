package vo

import (
	"time"
)

type SharedProjectInfo struct {
	Id          int    `json:"id"`
	OwnerId     string `json:"ownerId"`
	OwnerName   string `json:"ownerName"`
	ProjectName string `json:"projectName"`
	Description string `json:"description"`
}

type SessionUpdateReq struct {
	Title      *string `json:"title"`
	IsArchived *int    `json:"isArchived"`
}

type SessionListReq struct {
	Page      int  `url:"page"`
	PageSize  int  `url:"pageSize"`
	PersonaId *int `url:"personaId"`
}

type SessionListItem struct {
	SessionId     string             `json:"sessionId"`
	Title         string             `json:"title"`
	Status        uint8              `json:"status"`
	PersonaId     int                `json:"personaId"`
	Project       string             `json:"project"`
	SharedProject *SharedProjectInfo `json:"sharedProject,omitempty"`
	LastChatTime  time.Time          `json:"lastChatTime"`
	Created       time.Time          `json:"created"`
}

type SessionDetailRsp struct {
	SessionId             string             `json:"sessionId"`
	Title                 string             `json:"title"`
	Status                uint8              `json:"status"`
	PersonaId             int                `json:"personaId"`
	Project               string             `json:"project"`
	SharedProject         *SharedProjectInfo `json:"sharedProject,omitempty"`
	AIModelId             int                `json:"aiModelId"`
	ContextTokens         int                `json:"contextTokens"`
	PromptTokensCount     int                `json:"promptTokensCount" gorm:"column:prompt_tokens_count"`
	CompletionTokensCount int                `json:"completionTokensCount" gorm:"column:completion_tokens_count"`
	McpIds                []int              `json:"mcpIds"`
	SkillIds              []int              `json:"skillIds"`
	LastChatTime          time.Time          `json:"lastChatTime"`
	Created               time.Time          `json:"created"`

	ModelName    string `json:"modelName,omitempty"`
	PersonaName  string `json:"personaName,omitempty"`
	PersonaIcon  string `json:"personaIcon,omitempty"`
	ContextLimit int    `json:"contextLimit,omitempty"`
}

type MessageItem struct {
	MsgId            string                 `json:"msgId"`
	RoundId          string                 `json:"roundId"`
	ContextState     uint8                  `json:"contextState"`
	Content          string                 `json:"content"`
	ReasoningContent []MessageReasoningItem `json:"reasoningContent"`
	RoleType         string                 `json:"roleType"`
	Created          string                 `json:"created"`
}

type MessageReasoningToolCallsInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon"`
	Action      string `json:"action"`
}

type MessageReasoningItem struct {
	EventType string                         `json:"eventType"`
	Content   string                         `json:"content,omitempty"`
	Tool      *MessageReasoningToolCallsInfo `json:"tool,omitempty"`
}
