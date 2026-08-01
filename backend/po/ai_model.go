package po

import (
	"time"

	"gorm.io/gorm"
)

// AI 模型访问权限常量
const (
	AIModelAccessAll    = 0 // 全员开放
	AIModelAccessMember = 1 // 仅成员可见
)

// AIModel AI模型配置表
type AIModel struct {
	AIModelId           int       `gorm:"primaryKey;column:ai_model_id;type:int;autoIncrement"` // 主键ID
	ProviderDisplayName string    `gorm:"column:provider_display_name;type:varchar(128)"`       // 供应商显示名称，如 "DeepSeek"、"百炼"
	DisplayName         string    `gorm:"column:display_name;type:varchar(128)"`               // 模型显示名称，如 "DeepSeek V3"、"Qwen Turbo"
	ProviderID          string    `gorm:"index;column:provider_id;type:varchar(64)"`           // 供应商标识: deepseek, bailian, glm, minimax, open_router, ollama, openai, claude, gemini, openai_compatible, claude_compatible
	ModelName           string    `gorm:"column:model_name;type:varchar(128)"`                 // 模型名称，如 "deepseek-chat"
	Icon                string    `gorm:"column:icon;type:varchar(512)"`                      // 图标URL或Base64，非必填
	APIKey              string    `gorm:"column:api_key;type:text"`                            // API密钥，ollama可为空
	BaseURL             string    `gorm:"column:base_url;type:varchar(256)"`                   // 基础URL，ollama/openai_compatible/claude_compatible必填
	ExtraFields         string    `gorm:"column:extra_fields;type:text"`                       // 额外配置，JSON格式，如 {"thinking":{"type":"enabled"}}
	ProxyURL            string    `gorm:"column:proxy_url;type:varchar(256)"`                  // 代理地址，非必填，如 http://127.0.0.1:7890
	ContextLen          int       `gorm:"column:context_len;default:200"`                      // 上下文长度，单位KB，默认200，内部使用时需*1024
	IsDefault           uint8     `gorm:"column:is_default;default:0"`                         // 是否默认模型: 0否 1是（全局唯一）
	IsFlash             int       `gorm:"column:is_flash;default:0"`                          // 是否 Flash 模型: 0否 1是（全局唯一），用于历史对话压缩和子 agent
	IsVisual            int       `gorm:"column:is_visual;default:0"`                          // 是否多模态识别模型: 0否 1是（全局唯一），用于图片理解
	Status              uint8     `gorm:"column:status;default:1"`                            // 状态: 0禁用 1启用，禁用后会话中不出现该模型
	Access              uint8     `gorm:"column:access;default:0;not null"`                   // 访问权限：0全员开放 1仅成员可见
	Deleted             uint8     `gorm:"column:deleted;default:0"`                           // 软删除: 0正常 1删除
	Remark              string    `gorm:"column:remark;type:varchar(512)"`                    // 备注
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

// ContextLenInTokens 将KB转换为tokens数，供内部调用provider.CreateModel时使用
func (m *AIModel) ContextLenInTokens() int {
	return m.ContextLen * 1024
}
