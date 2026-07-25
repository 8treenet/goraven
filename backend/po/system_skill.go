package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	SystemSkillStatusDisabled = 1
	SystemSkillStatusEnabled  = 0
)

const SystemSkillNamePrefix = "goraven-"

type SystemSkill struct {
	SkillId     int       `gorm:"primaryKey;column:skill_id;type:int;autoIncrement"`
	Name        string    `gorm:"uniqueIndex;column:name;type:varchar(64);not null"`
	Description string    `gorm:"column:description;type:varchar(512)"`
	Content     string    `gorm:"column:content;type:text;not null"`
	Status      uint8     `gorm:"column:status;default:0"`
	Deleted     uint8     `gorm:"column:deleted;default:0"`
	Created     time.Time `gorm:"not null;column:created"`
	Updated     time.Time `gorm:"not null;column:updated"`
}

func (s *SystemSkill) TableName() string {
	return "system_skill"
}

func (s *SystemSkill) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

func (s *SystemSkill) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}
