package po

import (
	"time"

	"gorm.io/gorm"
)

type ToolDailyStats struct {
	StatId   int       `gorm:"primaryKey;column:id;autoIncrement"`
	UserId   string    `gorm:"uniqueIndex:idx_user_date_type_name;column:user_id;type:varchar(64);not null"`
	ToolType string    `gorm:"uniqueIndex:idx_user_date_type_name;index:idx_type_date_name,priority:1;column:tool_type;type:varchar(16);not null"`
	ToolName string    `gorm:"uniqueIndex:idx_user_date_type_name;index:idx_type_date_name,priority:3;column:tool_name;type:varchar(128);not null"`
	StatDate string    `gorm:"uniqueIndex:idx_user_date_type_name;index:idx_type_date_name,priority:2;column:stat_date;type:varchar(10);not null"`
	Count    int       `gorm:"column:count;default:0"`
	Created  time.Time `gorm:"not null;column:created"`
	Updated  time.Time `gorm:"not null;column:updated"`
}

func (s *ToolDailyStats) TableName() string {
	return "tool_daily_stats"
}

func (s *ToolDailyStats) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

func (s *ToolDailyStats) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}
