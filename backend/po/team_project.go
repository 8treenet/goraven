package po

import (
	"time"

	"gorm.io/gorm"
)

// 团队项目访问权限常量
const (
	TeamProjectAccessAll    = 0 // 全员开放
	TeamProjectAccessMember = 1 // 仅成员可见
)

// TeamProject 团队项目表
// 独立创建的团队项目，物理目录位于 team_project_dir/{ProjectName}，与个人工作空间完全解耦。
// ProjectName 全局唯一。
type TeamProject struct {
	Id             int        `gorm:"primaryKey;column:id;type:int;autoIncrement"`
	CreatorId      string     `gorm:"column:creator_id;type:varchar(64);index;not null"`          // 创建者 user_id
	ProjectName    string     `gorm:"column:project_name;type:varchar(255);uniqueIndex;not null"` // 项目目录名（全局唯一）
	Description    string     `gorm:"column:description;type:text"`                               // 项目简介，可空
	Access         uint8      `gorm:"column:access;default:0;not null"`                           // 访问权限：0全员开放 1仅成员可见
	VisitCount     int        `gorm:"column:visit_count;type:int;default:0;not null"`             // 访问次数
	LastActiveTime *time.Time `gorm:"column:last_active_time"`                                    // 最近活跃时间（空表示从未活跃）
	Created        time.Time  `gorm:"not null;column:created"`
	Updated        time.Time  `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (t *TeamProject) TableName() string {
	return "team_project"
}

// BeforeCreate 创建时设置时间戳
func (t *TeamProject) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.Created = now
	t.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (t *TeamProject) BeforeSave(tx *gorm.DB) error {
	t.Updated = time.Now()
	return nil
}
