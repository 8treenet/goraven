package po

import (
	"time"

	"gorm.io/gorm"
)

// PersonaCategory 角色分类表（管理员和用户角色共享）
type PersonaCategory struct {
	CategoryId int    `gorm:"primaryKey;column:category_id;type:int;autoIncrement"` // 主键ID
	Name       string `gorm:"column:name;type:varchar(128);not null"`              // 分类名称
	Icon       string `gorm:"column:icon;type:varchar(256)"`                       // 图标：Lucide 图标名称或 URL
	Deleted    uint8  `gorm:"column:deleted;default:0"`                            // 软删除：0 正常 1 删除
	IsDefault  uint8  `gorm:"column:is_default;default:0"`                          // 是否默认分类: 0否 1是
	Created    time.Time
	Updated    time.Time
}

func (c *PersonaCategory) TableName() string {
	return "persona_category"
}

func (c *PersonaCategory) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	c.Created = now
	c.Updated = now
	return nil
}

func (c *PersonaCategory) BeforeSave(tx *gorm.DB) error {
	c.Updated = time.Now()
	return nil
}
