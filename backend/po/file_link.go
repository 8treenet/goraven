package po

import (
	"time"

	"gorm.io/gorm"
)

type FileLink struct {
	LinkId		string		`gorm:"primaryKey;column:link_id;type:varchar(64)"`
	UserId		string		`gorm:"index;column:user_id;type:varchar(64)"`
	FilePath	string		`gorm:"column:file_path;type:varchar(512)"`
	FileName	string		`gorm:"column:file_name;type:varchar(255)"`
	ExpiresAt	time.Time	`gorm:"column:expires_at"`
	Deleted		uint8		`gorm:"column:deleted;default:0"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (f *FileLink) TableName() string {
	return "file_link"
}

func (f *FileLink) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	f.Created = now
	f.Updated = now
	return nil
}

func (f *FileLink) BeforeSave(tx *gorm.DB) error {
	f.Updated = time.Now()
	return nil
}

func (f *FileLink) IsExpired() bool {
	return time.Now().After(f.ExpiresAt)
}
