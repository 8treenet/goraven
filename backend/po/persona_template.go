package po

import (
	"time"

	"gorm.io/gorm"
)

type PersonaTemplate struct {
	TemplateId  int       `gorm:"primaryKey;column:template_id;type:int;autoIncrement"`
	Name        string    `gorm:"column:name;type:varchar(128);not null"`
	Description string    `gorm:"column:description;type:varchar(512)"`
	Icon        string    `gorm:"column:icon;type:varchar(256)"`
	RoleInfo    string    `gorm:"column:role_info;type:text"`
	CategoryId  int       `gorm:"column:category_id;type:int;default:0"`
	UsageCount  int       `gorm:"column:usage_count;type:int;default:0"`
	SortOrder   int       `gorm:"column:sort_order;default:0"`
	Deleted     uint8     `gorm:"column:deleted;default:0"`
	Created     time.Time `gorm:"not null;column:created"`
	Updated     time.Time `gorm:"not null;column:updated"`
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
