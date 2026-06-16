package repository

import (
	"raven/backend/po"
	"raven/util"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *DailyStatsRepository {
			return &DailyStatsRepository{}
		})
	})
}

type DailyStatsRepository struct {
	freedom.Repository
}

func (repo *DailyStatsRepository) AddDailyStats(userId string, promptTokens, completionTokens, promptCachedTokens int) error {
	today := util.StartOfToday().Format("2006-01-02")
	stats := &po.UserDailyStats{
		UserId:			userId,
		StatDate:		today,
		PromptTokens:		promptTokens,
		CompletionTokens:	completionTokens,
		PromptCachedTokens:	promptCachedTokens,
		MessageCount:		1,
		RoundCount:		1,
	}
	return repo.db().Clauses(clause.OnConflict{
		Columns:	[]clause.Column{{Name: "user_id"}, {Name: "stat_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"prompt_tokens":	gorm.Expr("user_daily_stats.prompt_tokens + ?", promptTokens),
			"completion_tokens":	gorm.Expr("user_daily_stats.completion_tokens + ?", completionTokens),
			"prompt_cached_tokens":	gorm.Expr("user_daily_stats.prompt_cached_tokens + ?", promptCachedTokens),
			"message_count":	gorm.Expr("user_daily_stats.message_count + 1"),
			"round_count":		gorm.Expr("user_daily_stats.round_count + 1"),
			"updated":		time.Now(),
		}),
	}).Create(stats).Error
}

func (repo *DailyStatsRepository) AddToolDailyStats(userId, toolType, toolName string) error {
	today := util.StartOfToday().Format("2006-01-02")
	stats := &po.ToolDailyStats{
		UserId:		userId,
		ToolType:	toolType,
		ToolName:	toolName,
		StatDate:	today,
		Count:		1,
	}
	return repo.db().Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "tool_type"},
			{Name: "tool_name"},
			{Name: "stat_date"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"count":	gorm.Expr("tool_daily_stats.count + ?", 1),
			"updated":	time.Now(),
		}),
	}).Create(stats).Error
}

func (repo *DailyStatsRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
