package po

import (
	"time"

	"gorm.io/gorm"
)

// TeamProjectMember 团队项目成员表
// 记录用户与团队项目的归属关系。
// 同一用户在同一项目中仅允许一条记录（联合唯一索引）。
type TeamProjectMember struct {
	Id        int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`
	ProjectId int       `gorm:"column:project_id;type:int;uniqueIndex:uk_project_user;not null"`      // 关联 team_project.id
	UserId    string    `gorm:"column:user_id;type:varchar(64);uniqueIndex:uk_project_user;not null"` // 关联 user.user_id
	Created   time.Time `gorm:"not null;column:created"`
	Updated   time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (t *TeamProjectMember) TableName() string {
	return "team_project_member"
}

// BeforeCreate 创建时设置时间戳
func (t *TeamProjectMember) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.Created = now
	t.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (t *TeamProjectMember) BeforeSave(tx *gorm.DB) error {
	t.Updated = time.Now()
	return nil
}
