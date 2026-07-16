package po

import (
	"time"

	"gorm.io/gorm"
)

type SharedProject struct {
	Id           int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`
	OwnerId      string    `gorm:"column:owner_id;type:varchar(64);index;uniqueIndex:idx_owner_project;not null"`
	ProjectName  string    `gorm:"column:project_name;type:varchar(255);uniqueIndex:idx_owner_project;not null"`
	Description  string    `gorm:"column:description;type:text"`
	VisitCount   int       `gorm:"column:visit_count;type:int;default:0;not null"`
	LastActiveAt time.Time `gorm:"column:last_active_at"`
	Created      time.Time `gorm:"not null;column:created"`
	Updated      time.Time `gorm:"not null;column:updated"`
}

func (s *SharedProject) TableName() string {
	return "shared_project"
}

func (s *SharedProject) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

func (s *SharedProject) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}
