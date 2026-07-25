package po

import (
	"time"

	"gorm.io/gorm"
)

type AIModel struct {
	AIModelId           int       `gorm:"primaryKey;column:ai_model_id;type:int;autoIncrement"`
	ProviderDisplayName string    `gorm:"column:provider_display_name;type:varchar(128)"`
	DisplayName         string    `gorm:"column:display_name;type:varchar(128)"`
	ProviderID          string    `gorm:"index;column:provider_id;type:varchar(64)"`
	ModelName           string    `gorm:"column:model_name;type:varchar(128)"`
	Icon                string    `gorm:"column:icon;type:varchar(512)"`
	APIKey              string    `gorm:"column:api_key;type:text"`
	BaseURL             string    `gorm:"column:base_url;type:varchar(256)"`
	ExtraFields         string    `gorm:"column:extra_fields;type:text"`
	ProxyURL            string    `gorm:"column:proxy_url;type:varchar(256)"`
	ContextLen          int       `gorm:"column:context_len;default:200"`
	IsDefault           uint8     `gorm:"column:is_default;default:0"`
	IsFlash             int       `gorm:"column:is_flash;default:0"`
	IsVisual            int       `gorm:"column:is_visual;default:0"`
	Status              uint8     `gorm:"column:status;default:1"`
	Deleted             uint8     `gorm:"column:deleted;default:0"`
	Remark              string    `gorm:"column:remark;type:varchar(512)"`
	Created             time.Time `gorm:"not null;column:created"`
	Updated             time.Time `gorm:"not null;column:updated"`
}

func (m *AIModel) TableName() string {
	return "ai_model"
}

func (m *AIModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.Created = now
	m.Updated = now
	if m.ContextLen <= 0 {
		m.ContextLen = 200
	}
	return nil
}

func (m *AIModel) BeforeSave(tx *gorm.DB) error {
	m.Updated = time.Now()
	if m.ContextLen <= 0 {
		m.ContextLen = 200
	}
	return nil
}

func (m *AIModel) ContextLenInTokens() int {
	return m.ContextLen * 1024
}
