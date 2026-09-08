package po

import (
	"time"

	"gorm.io/gorm"
)

// UserProject stores a user's personal project metadata.
type UserProject struct {
	Id          int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`
	UserId      string    `gorm:"column:user_id;type:varchar(64);uniqueIndex:idx_user_project;not null"`
	ProjectName string    `gorm:"column:project_name;type:varchar(255);uniqueIndex:idx_user_project;not null"`
	Description string    `gorm:"column:description;type:text"`
	GitUrl      string    `gorm:"column:git_url;type:varchar(2048)"`
	Created     time.Time `gorm:"not null;column:created"`
	Updated     time.Time `gorm:"not null;column:updated;autoUpdateTime"`
}

// TableName sets the database table name.
func (p *UserProject) TableName() string {
	return "user_project"
}

// BeforeCreate sets both timestamps when creating a project.
func (p *UserProject) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.Created = now
	p.Updated = now
	return nil
}

// BeforeSave refreshes the update timestamp.
func (p *UserProject) BeforeSave(tx *gorm.DB) error {
	p.Updated = time.Now()
	return nil
}
