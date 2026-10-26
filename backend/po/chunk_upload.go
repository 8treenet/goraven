package po

import (
	"time"

	"gorm.io/gorm"
)

// 上传状态常量
const (
	UploadStatusPending   uint8 = 0 // 进行中
	UploadStatusCompleted uint8 = 1 // 已完成（已合并）
	UploadStatusCancelled uint8 = 2 // 已取消
	UploadStatusUsed      uint8 = 3 // 已使用（已转为静态资源等）
)

// ChunkUpload 分片上传任务表
// 记录用户分片上传的任务状态，支持断点续传
type ChunkUpload struct {
	UploadId    string    `gorm:"primaryKey;column:upload_id;type:varchar(64)"` // 上传任务唯一标识
	UserId      string    `gorm:"index;column:user_id;type:varchar(64)"`        // 用户ID
	FileName    string    `gorm:"column:file_name;type:varchar(255)"`           // 原始文件名
	FileSize    int64     `gorm:"column:file_size"`                             // 文件总大小（字节）
	ChunkSize   int       `gorm:"column:chunk_size"`                            // 分片大小（字节）
	TotalChunks int       `gorm:"column:total_chunks"`                          // 总分片数
	TempDir     string    `gorm:"column:temp_dir;type:varchar(512)"`            // 临时存储目录
	Status      uint8     `gorm:"column:status;default:0;index"`               // 状态：0进行中 1已完成 2已取消 3已使用
	Deleted     uint8     `gorm:"column:deleted;default:0;index"`              // 软删除：0正常 1已删除
	Created     time.Time `gorm:"not null;column:created"`
	Updated     time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (c *ChunkUpload) TableName() string {
	return "chunk_upload"
}

// BeforeCreate 创建时设置时间戳
func (c *ChunkUpload) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	c.Created = now
	c.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (c *ChunkUpload) BeforeSave(tx *gorm.DB) error {
	c.Updated = time.Now()
	return nil
}
