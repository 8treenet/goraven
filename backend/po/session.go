package po

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	SessionId             string    `gorm:"primaryKey;column:session_id;type:varchar(64)"`
	UserId                string    `gorm:"index;index:idx_user_del_arch,priority:1;column:user_id"`
	Title                 string    `gorm:"column:title;type:varchar(255)"`
	IsArchived            int       `gorm:"index:idx_user_del_arch,priority:3;column:is_archived"`
	PromptTokensCount     int       `gorm:"column:prompt_tokens_count"`
	CompletionTokensCount int       `gorm:"column:completion_tokens_count"`
	PromptCachedTokens    int       `gorm:"column:prompt_cached_tokens;default:0"`
	ContextTokens         int       `gorm:"column:context_tokens"`
	Status                uint8     `gorm:"column:status"`
	AIModelId             int       `gorm:"index;column:ai_model_id"`
	LastChatTime          time.Time `gorm:"not null;column:last_chat_time"`
	PersonaId             int       `gorm:"column:persona_id;index;default:0"`
	McpIds                string    `gorm:"column:mcp_ids;type:text"`
	SkillIds              string    `gorm:"column:skill_ids;type:text"`
	Project               string    `gorm:"column:project;type:varchar(255);default:''"`
	SharedProjectId       int       `gorm:"column:shared_project_id;index;default:0"`
	Deleted               uint8     `gorm:"index:idx_user_del_arch,priority:2;index:idx_del_created,priority:1;column:deleted;default:0"`
	Created               time.Time `gorm:"index:idx_del_created,priority:2;not null;column:created"`
	Updated               time.Time `gorm:"not null;column:updated"`
}

func (session *Session) TableName() string {
	return "session"
}

func (session *Session) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	session.Created = now
	session.Updated = now
	session.LastChatTime = now
	return nil
}

func (session *Session) BeforeSave(tx *gorm.DB) error {
	session.Updated = time.Now()
	return nil
}
