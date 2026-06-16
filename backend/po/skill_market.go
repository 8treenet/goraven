package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	SkillSourceClawHub	= "clawhub"
	SkillSourceCustomUpload	= "custom_upload"
	SkillSourceSystem	= "system"
)

const (
	SkillStatusDisabled	= 0
	SkillStatusEnabled	= 1
)

type SkillMarket struct {
	SkillId		int	`gorm:"primaryKey;column:skill_id;type:int;autoIncrement"`
	Name		string	`gorm:"uniqueIndex;column:name;type:varchar(64);not null"`
	Description	string	`gorm:"column:description;type:varchar(800)"`
	Icon		string	`gorm:"column:icon;type:varchar(256)"`

	Source		string	`gorm:"column:source;type:varchar(16)"`
	SourceUrl	string	`gorm:"column:source_url;type:varchar(512)"`
	CategoryId	int	`gorm:"column:category_id;index"`

	Status		uint8		`gorm:"column:status;default:1"`
	SortOrder	int		`gorm:"column:sort_order;default:0"`
	InstalledCount	int		`gorm:"column:installed_count;default:0"`
	Remark		string		`gorm:"column:remark;type:varchar(512)"`
	Deleted		uint8		`gorm:"column:deleted;default:0"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (s *SkillMarket) TableName() string {
	return "skill_market"
}

func (s *SkillMarket) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

func (s *SkillMarket) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

func (s *SkillMarket) IsEnabled() bool {
	return s.Status == SkillStatusEnabled && s.Deleted == 0
}
