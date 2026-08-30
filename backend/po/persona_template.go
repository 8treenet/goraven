package po

import (
	"time"

	"gorm.io/gorm"
)

// PersonaTemplate 角色模板表（管理员预设的系统提示词模板，用户选择后仅预填内容，无强绑定）
type PersonaTemplate struct {
	TemplateId        int       `gorm:"primaryKey;column:template_id;type:int;autoIncrement"`  // 主键ID
	Name              string    `gorm:"column:name;type:varchar(128);not null"`                // 模板名称（展示名称）
	Description       string    `gorm:"column:description;type:varchar(512)"`                  // 模板描述
	Icon              string    `gorm:"column:icon;type:varchar(256)"`                         // 图标：Lucide 图标名称或 URL
	RoleInfo          string    `gorm:"column:role_info;type:text"`                             // 系统提示词（核心内容）
	CategoryId        int       `gorm:"column:category_id;type:int;default:0"`                  // 分类ID，关联 persona_category 表
	UsageCount        int       `gorm:"column:usage_count;type:int;default:0"`                  // 使用次数（冗余计数，仅增不减）
	SortOrder         int       `gorm:"column:sort_order;default:0"`                            // 排序
	Deleted           uint8     `gorm:"column:deleted;default:0"`                              // 软删除：0 正常 1 删除
	Created           time.Time `gorm:"not null;column:created"`
	Updated           time.Time `gorm:"not null;column:updated"`
}

func (p *PersonaTemplate) TableName() string {
	return "persona_template"
}

func (p *PersonaTemplate) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.Created = now
	p.Updated = now
	return nil
}

func (p *PersonaTemplate) BeforeSave(tx *gorm.DB) error {
	p.Updated = time.Now()
	return nil
}
