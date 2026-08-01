package po

import (
	"time"

	"gorm.io/gorm"
)

// AIModelMember 模型可见成员表
// 记录用户与模型的可见性归属关系。
// 同一用户在同一模型下仅允许一条记录（联合唯一索引）。
type AIModelMember struct {
	Id        int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`
	AIModelId int       `gorm:"column:ai_model_id;type:int;uniqueIndex:uk_model_user;not null"`      // 关联 ai_model.ai_model_id
	UserId    string    `gorm:"column:user_id;type:varchar(64);uniqueIndex:uk_model_user;not null"` // 关联 user.user_id
	Created   time.Time `gorm:"not null;column:created"`
	Updated   time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (m *AIModelMember) TableName() string {
	return "ai_model_member"
}

// BeforeCreate 创建时设置时间戳
func (m *AIModelMember) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.Created = now
	m.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (m *AIModelMember) BeforeSave(tx *gorm.DB) error {
	m.Updated = time.Now()
	return nil
}
