package po

import (
	"time"

	"gorm.io/gorm"
)

// UserDailyStats 用户日统计表
// 按 (用户, 日期) 粒度预聚合 Token 消耗、消息数、轮次等数据
// 每轮对话 SSE end 事件后增量更新，避免仪表盘实时聚合 message 表
type UserDailyStats struct {
	StatId             int       `gorm:"primaryKey;column:id;autoIncrement"`
	UserId             string    `gorm:"uniqueIndex:idx_user_date;column:user_id;type:varchar(64);not null"` // 用户ID
	StatDate           string    `gorm:"uniqueIndex:idx_user_date;index;column:stat_date;type:varchar(10);not null"` // 统计日期 (YYYY-MM-DD)
	PromptTokens       int       `gorm:"column:prompt_tokens;default:0"`                                     // 当日 prompt token
	PromptCachedTokens int       `gorm:"column:prompt_cached_tokens;default:0"`                               // 当日 缓存 prompt token
	CompletionTokens   int       `gorm:"column:completion_tokens;default:0"`                                 // 当日 completion token
	MessageCount       int       `gorm:"column:message_count;default:0"`                                     // 当日消息数
	RoundCount         int       `gorm:"column:round_count;default:0"`                                       // 当日对话轮次
	Created            time.Time `gorm:"not null;column:created"`
	Updated            time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (s *UserDailyStats) TableName() string {
	return "user_daily_stats"
}

// BeforeCreate 创建时设置时间戳
func (s *UserDailyStats) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (s *UserDailyStats) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

// TotalTokens 返回当日总 token
func (s *UserDailyStats) TotalTokens() int {
	return s.PromptTokens + s.CompletionTokens
}
