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

// SystemSkill 系统技能表（管理员维护的全局提示词技能）
type SystemSkill struct {
	SkillId     int       `gorm:"primaryKey;column:skill_id;type:int;autoIncrement"`  // 主键ID
	Name        string    `gorm:"uniqueIndex;column:name;type:varchar(64);not null"`       // 唯一标识名，必须是goraven-开头
	Description string    `gorm:"column:description;type:varchar(512)"`              // 简短描述
	Content     string    `gorm:"column:content;type:text;not null"`                 // 技能内容（提示词）
	Version     string    `gorm:"column:version;type:varchar(32)"`                   // 内容版本号（对应 config.Version，版本升级时自动更新内容）
	Status      uint8     `gorm:"column:status;default:0"`                           // 状态：0启用 1禁用
	Deleted     uint8     `gorm:"column:deleted;default:0"`                          // 软删除：0正常 1删除
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
