package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	UserSkillInstallPending		= 0
	UserSkillInstallProgress	= 1
	UserSkillInstalled		= 2
	UserSkillInstallFailed		= 3
)

type UserSkill struct {
	UserSkillId	int	`gorm:"primaryKey;column:user_skill_id;type:int;autoIncrement"`
	UserId		string	`gorm:"uniqueIndex:idx_user_skill;column:user_id;type:varchar(64);not null"`
	SkillName	string	`gorm:"uniqueIndex:idx_user_skill;column:skill_name;type:varchar(64);not null"`
	Description	string	`gorm:"column:description;type:varchar(800)"`
	Icon		string	`gorm:"column:icon;type:varchar(256)"`

	MarketSkillId	int	`gorm:"column:market_skill_id;index"`
	CategoryId	int	`gorm:"column:category_id;index"`
	Source		string	`gorm:"column:source;type:varchar(20)"`
	InstallStatus	uint8	`gorm:"column:install_status;default:0"`
	InstallError	string	`gorm:"column:install_error;type:text"`
	AlwaysOn	uint8	`gorm:"column:always_on;default:0"`

	Created	time.Time	`gorm:"not null;column:created"`
	Updated	time.Time	`gorm:"not null;column:updated"`
}

func (us *UserSkill) TableName() string {
	return "user_skill"
}

func (us *UserSkill) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	us.Created = now
	us.Updated = now
	return nil
}

func (us *UserSkill) BeforeSave(tx *gorm.DB) error {
	us.Updated = time.Now()
	return nil
}

func (us *UserSkill) IsInstalled() bool {
	return us.InstallStatus == UserSkillInstalled
}
