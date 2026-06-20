package po

import (
	"time"

	"gorm.io/gorm"
)

type SkillShare struct {
	ShareId      int       `gorm:"primaryKey;column:share_id;type:int;autoIncrement"`
	OwnerId      string    `gorm:"column:owner_id;type:varchar(64);index;not null"`
	SkillName    string    `gorm:"column:skill_name;type:varchar(64);uniqueIndex;not null"`
	Description  string    `gorm:"column:description;type:varchar(800)"`
	Icon         string    `gorm:"column:icon;type:varchar(256)"`
	CategoryId   int       `gorm:"column:category_id"`
	Note         string    `gorm:"column:note;type:varchar(800)"`
	InstallCount int       `gorm:"column:install_count;default:0"`
	Created      time.Time `gorm:"not null;column:created"`
	Updated      time.Time `gorm:"not null;column:updated"`
}

func (s *SkillShare) TableName() string {
	return "skill_share"
}

func (s *SkillShare) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

func (s *SkillShare) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}
