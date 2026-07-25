package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	UploadStatusPending   uint8 = 0
	UploadStatusCompleted uint8 = 1
	UploadStatusCancelled uint8 = 2
	UploadStatusUsed      uint8 = 3
)

type ChunkUpload struct {
	UploadId    string    `gorm:"primaryKey;column:upload_id;type:varchar(64)"`
	UserId      string    `gorm:"index;column:user_id;type:varchar(64)"`
	FileName    string    `gorm:"column:file_name;type:varchar(255)"`
	FileSize    int64     `gorm:"column:file_size"`
	ChunkSize   int       `gorm:"column:chunk_size"`
	TotalChunks int       `gorm:"column:total_chunks"`
	TempDir     string    `gorm:"column:temp_dir;type:varchar(512)"`
	Status      uint8     `gorm:"column:status;default:0;index"`
	Deleted     uint8     `gorm:"column:deleted;default:0;index"`
	Created     time.Time `gorm:"not null;column:created"`
	Updated     time.Time `gorm:"not null;column:updated"`
}

func (c *ChunkUpload) TableName() string {
	return "chunk_upload"
}

func (c *ChunkUpload) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	c.Created = now
	c.Updated = now
	return nil
}

func (c *ChunkUpload) BeforeSave(tx *gorm.DB) error {
	c.Updated = time.Now()
	return nil
}
