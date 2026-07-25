package po

import (
	"time"

	"gorm.io/gorm"
)

const RoleTypeUser = "user"
const RoleTypeAssistant = "assistant"
const RoleTypeSummary = "summary"
const RoleTypeTool = "tool"

type ToolCallData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolInfo struct {
	ToolCallID string `json:"toolCallId,omitempty"`
	ToolName   string `json:"toolName,omitempty"`
}

type AssistantExt struct {
	ToolCalls []ToolCallData `json:"toolCalls,omitempty"`
}

type ToolExt struct {
	ToolCallID string `json:"toolCallId"`
	ToolName   string `json:"toolName"`
}

type Message struct {
	MsgId                 string    `gorm:"primaryKey;column:msg_id;type:varchar(64)"`
	SessionId             string    `gorm:"index;index:idx_session_round,priority:1;column:session_id;type:varchar(64)"`
	RoundId               string    `gorm:"index:idx_session_round,priority:2;column:round_id;type:varchar(64)"`
	Timestamp             int64     `gorm:"index;column:timestamp"`
	ContextState          uint8     `gorm:"column:context_state"`
	Content               string    `gorm:"column:content;type:text"`
	ReasoningContent      string    `gorm:"column:reasoning_content;type:text"`
	RoleType              string    `gorm:"column:role_type;type:varchar(32)"`
	Tool                  uint8     `gorm:"column:tool"`
	PromptTokensCount     int       `gorm:"column:prompt_tokens_count"`
	CompletionTokensCount int       `gorm:"column:completion_tokens_count"`
	Duration              int       `gorm:"column:duration"`
	AsstError             string    `gorm:"column:asst_error;type:text"`
	Ext                   string    `gorm:"column:ext;type:text"`
	ToolCallsInfo         string    `gorm:"column:tool_calls_info;type:text"`
	Created               time.Time `gorm:"not null;column:created"`
	Updated               time.Time `gorm:"not null;column:updated"`
}

func (msg *Message) TableName() string {
	return "message"
}

func (msg *Message) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	msg.Created = now
	msg.Updated = now
	return nil
}

func (msg *Message) BeforeSave(tx *gorm.DB) error {
	msg.Updated = time.Now()
	return nil
}
