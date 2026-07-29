package po

import (
	"time"

	"gorm.io/gorm"
)

type TeamProject struct {
	Id		int		`gorm:"primaryKey;column:id;type:int;autoIncrement"`
	CreatorId	string		`gorm:"column:creator_id;type:varchar(64);index;not null"`
	ProjectName	string		`gorm:"column:project_name;type:varchar(255);uniqueIndex;not null"`
	Description	string		`gorm:"column:description;type:text"`
	VisitCount	int		`gorm:"column:visit_count;type:int;default:0;not null"`
	LastActiveAt	time.Time	`gorm:"column:last_active_at"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (t *TeamProject) TableName() string {
	return "team_project"
}

func (t *TeamProject) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.Created = now
	t.Updated = now
	return nil
}

func (t *TeamProject) BeforeSave(tx *gorm.DB) error {
	t.Updated = time.Now()
	return nil
}
