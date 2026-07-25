package po

import (
	"time"

	"gorm.io/gorm"
)

type SkillCategory struct {
	CategoryId int    `gorm:"primaryKey;column:category_id;type:int;autoIncrement"`
	Name       string `gorm:"column:name;type:varchar(128);not null"`
	Icon       string `gorm:"column:icon;type:varchar(256)"`
	Deleted    uint8  `gorm:"column:deleted;default:0"`
	IsDefault  uint8  `gorm:"column:is_default;default:0"`
	Created    time.Time
	Updated    time.Time
}

func (c *SkillCategory) TableName() string {
	return "skill_category"
}

func (c *SkillCategory) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	c.Created = now
	c.Updated = now
	return nil
}

func (c *SkillCategory) BeforeSave(tx *gorm.DB) error {
	c.Updated = time.Now()
	return nil
}
