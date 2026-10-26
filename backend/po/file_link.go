package po

import (
	"time"

	"gorm.io/gorm"
)

// FileLink 文件外链映射表
// 存储用户文件与外部可访问链接ID的映射关系
type FileLink struct {
	LinkId    string    `gorm:"primaryKey;column:link_id;type:varchar(64)"` // 外链唯一标识，用于URL中的文件名
	UserId    string    `gorm:"index;column:user_id;type:varchar(64)"`      // 用户ID
	FilePath  string    `gorm:"column:file_path;type:varchar(512)"`         // 用户目录下的相对路径，如 /temp/screenshot.png
	FileName  string    `gorm:"column:file_name;type:varchar(255)"`         // 原始文件名，用于下载时展示
	ExpiresAt time.Time `gorm:"column:expires_at"`                          // 失效时间
	Deleted   uint8     `gorm:"column:deleted;default:0"`                  // 软删除: 0正常 1删除
	Created   time.Time `gorm:"not null;column:created"`
	Updated   time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (f *FileLink) TableName() string {
	return "file_link"
}

// BeforeCreate 创建时设置时间戳
func (f *FileLink) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	f.Created = now
	f.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (f *FileLink) BeforeSave(tx *gorm.DB) error {
	f.Updated = time.Now()
	return nil
}

// IsExpired 检查是否已过期
func (f *FileLink) IsExpired() bool {
	return time.Now().After(f.ExpiresAt)
}
