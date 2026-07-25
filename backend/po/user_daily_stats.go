package po

import (
	"time"

	"gorm.io/gorm"
)

type UserDailyStats struct {
	StatId             int       `gorm:"primaryKey;column:id;autoIncrement"`
	UserId             string    `gorm:"uniqueIndex:idx_user_date;column:user_id;type:varchar(64);not null"`
	StatDate           string    `gorm:"uniqueIndex:idx_user_date;index;column:stat_date;type:varchar(10);not null"`
	PromptTokens       int       `gorm:"column:prompt_tokens;default:0"`
	PromptCachedTokens int       `gorm:"column:prompt_cached_tokens;default:0"`
	CompletionTokens   int       `gorm:"column:completion_tokens;default:0"`
	MessageCount       int       `gorm:"column:message_count;default:0"`
	RoundCount         int       `gorm:"column:round_count;default:0"`
	Created            time.Time `gorm:"not null;column:created"`
	Updated            time.Time `gorm:"not null;column:updated"`
}

func (s *UserDailyStats) TableName() string {
	return "user_daily_stats"
}

func (s *UserDailyStats) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

func (s *UserDailyStats) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

func (s *UserDailyStats) TotalTokens() int {
	return s.PromptTokens + s.CompletionTokens
}
