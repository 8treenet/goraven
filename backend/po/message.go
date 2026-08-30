package po

import (
	"time"

	"gorm.io/gorm"
)

const RoleTypeUser = "user"
const RoleTypeAssistant = "assistant"
const RoleTypeSummary = "summary"
const RoleTypeTool = "tool" // 工具消息

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
	MsgId                   string    `gorm:"primaryKey;column:msg_id;type:varchar(64)"`                                   //主键唯一id
	SessionId               string    `gorm:"index;index:idx_session_round,priority:1;column:session_id;type:varchar(64)"` //会话id，需要索引
	RoundId                 string    `gorm:"index:idx_session_round,priority:2;column:round_id;type:varchar(64)"`         //轮次id，同一轮对话（问题、回复、工具消息）共享
	Timestamp               int64     `gorm:"index;column:timestamp"`                                                      //时间戳，正序的历史记录
	ContextState            uint8     `gorm:"column:context_state"`                                                        //上下文状态: 0=发送给LLM, 1=跳过上下文(仅展示给用户)
	Content                 string    `gorm:"column:content;type:text"`                                                    //消息内容
	ReasoningContent        string    `gorm:"column:reasoning_content;type:text"`                                          //推理内容
	RoleType                string    `gorm:"column:role_type;type:varchar(32)"`                                           //角色 user：用户发出的， assistant：ai返回的, summary:压缩的摘要信息, tool:工具消息
	Tool                    uint8     `gorm:"column:tool"`                                                                 //0正常对话消息,1工具调用消息
	PromptTokensCount       int       `gorm:"column:prompt_tokens_count"`                                                  //本轮prompt token用量；仅每轮最后一条assistant消息（或压缩产生的summary消息）记录，值为整轮所有LLM调用的累计
	CompletionTokensCount   int       `gorm:"column:completion_tokens_count"`                                              //本轮completion token用量；同上，仅轮次最后一条assistant/summary消息记录整轮累计值
	PromptCachedTokensCount int       `gorm:"column:prompt_cached_tokens_count"`                                           //本轮缓存命中的prompt token；同上，仅轮次最后一条assistant/summary消息记录整轮累计值
	Duration                int       `gorm:"column:duration"`                                                             //耗时 毫秒
	AsstError               string    `gorm:"column:asst_error;type:text"`                                                 //助理的回复如果失败,记录原因
	Ext                     string    `gorm:"column:ext;type:text"`                                                        //扩展数据, assistant消息存AssistantExt, tool消息存ToolExt
	ToolCallsInfo           string    `gorm:"column:tool_calls_info;type:text"`
	Created                 time.Time `gorm:"not null;column:created"`
	Updated                 time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (msg *Message) TableName() string {
	return "message"
}

// BeforeCreate 创建时设置时间戳
func (msg *Message) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	msg.Created = now
	msg.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (msg *Message) BeforeSave(tx *gorm.DB) error {
	msg.Updated = time.Now()
	return nil
}
