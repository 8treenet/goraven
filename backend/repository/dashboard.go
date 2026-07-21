package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"raven/backend/infra"
	"raven/backend/po"
	"raven/backend/vo"
	"raven/config"

	"github.com/8treenet/freedom"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *DashboardRepository {
			return &DashboardRepository{}
		})
	})
}

type DashboardRepository struct {
	freedom.Repository
}

func (repo *DashboardRepository) quoteUser() string {
	if config.Get().System.DBType == "mysql" {
		return "`user`"
	}
	return `"user"`
}

func (repo *DashboardRepository) dateExpr() string {
	switch config.Get().System.DBType {
	case "mysql":
		return "DATE(FROM_UNIXTIME(timestamp/1000))"
	case "pg":
		return "DATE(TO_TIMESTAMP(timestamp/1000.0))"
	default:
		return "date(timestamp/1000, 'unixepoch')"
	}
}

func (repo *DashboardRepository) GetActiveUsersInRange(startDaysAgo, endDaysAgo int) (int64, error) {
	dates := genDateRange(startDaysAgo-1, endDaysAgo)
	if len(dates) == 0 {
		return 0, nil
	}

	var count int64
	err := repo.db().
		Table("message").
		Select("COUNT(DISTINCT session.user_id)").
		Joins("JOIN session ON message.session_id = session.session_id").
		Where(fmt.Sprintf("%s IN ?", repo.dateExpr()), dates).
		Scan(&count).Error
	return count, err
}

func (repo *DashboardRepository) GetSessionCounts() (total int64, newThisWeek int64, err error) {
	if err = repo.db().Model(&po.Session{}).Where("deleted = 0").Count(&total).Error; err != nil {
		return 0, 0, err
	}

	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	mondayStart := time.Date(now.Year(), now.Month(), now.Day()-int(weekday)+1, 0, 0, 0, 0, now.Location())

	if err = repo.db().Model(&po.Session{}).Where("deleted = 0 AND created >= ?", mondayStart).Count(&newThisWeek).Error; err != nil {
		return 0, 0, err
	}
	return total, newThisWeek, nil
}

func (repo *DashboardRepository) GetTodayTokens() (int64, error) {
	today := time.Now().Format("2006-01-02")
	var total int64
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0)").
		Where("stat_date = ?", today).
		Scan(&total).Error
	return total, err
}

func (repo *DashboardRepository) GetEnabledModelCount() (int64, error) {
	var count int64
	err := repo.db().Model(&po.AIModel{}).Where("status = 1 AND deleted = 0").Count(&count).Error
	return count, err
}

func (repo *DashboardRepository) GetTokensInDateRange(dateRange []string) (prompt int64, completion int64, err error) {
	type result struct {
		Prompt     int64
		Completion int64
	}
	var r result
	err = repo.db().
		Model(&po.UserDailyStats{}).
		Select("COALESCE(SUM(prompt_tokens), 0) AS prompt, COALESCE(SUM(completion_tokens), 0) AS completion").
		Where("stat_date IN ?", dateRange).
		Scan(&r).Error
	return r.Prompt, r.Completion, err
}

func (repo *DashboardRepository) GetSparkline(dateRange []string) ([]vo.SparklineItem, error) {
	var rows []struct {
		StatDate string `gorm:"column:stat_date"`
		Tokens   int64  `gorm:"column:tokens"`
	}
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("stat_date, SUM(prompt_tokens + completion_tokens) AS tokens").
		Where("stat_date IN ?", dateRange).
		Group("stat_date").
		Order("stat_date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]vo.SparklineItem, 0, len(dateRange))
	dateMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		dateMap[r.StatDate] = r.Tokens
	}
	for _, d := range dateRange {
		result = append(result, vo.SparklineItem{Date: d, Tokens: dateMap[d]})
	}
	return result, nil
}

func (repo *DashboardRepository) GetTokenTrend(dateRange []string) ([]vo.TokenTrendItem, error) {
	var rows []struct {
		StatDate         string `gorm:"column:stat_date"`
		PromptTokens     int64  `gorm:"column:prompt_tokens"`
		CompletionTokens int64  `gorm:"column:completion_tokens"`
	}
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("stat_date, COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens, COALESCE(SUM(completion_tokens), 0) AS completion_tokens").
		Where("stat_date IN ?", dateRange).
		Group("stat_date").
		Order("stat_date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	dateMap := make(map[string]*vo.TokenTrendItem, len(rows))
	for i := range rows {
		ds := rows[i].StatDate
		dateMap[ds] = &vo.TokenTrendItem{
			Date:             ds,
			PromptTokens:     rows[i].PromptTokens,
			CompletionTokens: rows[i].CompletionTokens,
		}
	}

	result := make([]vo.TokenTrendItem, 0, len(dateRange))
	for _, d := range dateRange {
		if item, ok := dateMap[d]; ok {
			result = append(result, *item)
		} else {
			result = append(result, vo.TokenTrendItem{Date: d})
		}
	}
	return result, nil
}

func (repo *DashboardRepository) GetModelUsage(dateRange []string) ([]vo.ModelUsageItem, error) {
	deletedLabel := "已删除模型"
	if config.Get().GetLanguage() == "en" {
		deletedLabel = "Deleted Model"
	}

	type row struct {
		ModelName        string
		Tokens           int64
		PromptTokens     int64
		CompletionTokens int64
	}

	query := repo.db().
		Table("session").
		Select(fmt.Sprintf("CASE WHEN ai_model.ai_model_id IS NULL OR ai_model.deleted = 1 THEN '%s' ELSE ai_model.display_name END AS model_name, SUM(session.prompt_tokens_count + session.completion_tokens_count) AS tokens, SUM(session.prompt_tokens_count) AS prompt_tokens, SUM(session.completion_tokens_count) AS completion_tokens", deletedLabel)).
		Joins("LEFT JOIN ai_model ON session.ai_model_id = ai_model.ai_model_id").
		Where("session.deleted = 0").
		Group(fmt.Sprintf("CASE WHEN ai_model.ai_model_id IS NULL OR ai_model.deleted = 1 THEN '%s' ELSE ai_model.display_name END", deletedLabel))

	if len(dateRange) > 0 {
		startDate, endDate := parseDateRange(dateRange)
		query = query.Where("session.created >= ? AND session.created < ?", startDate, endDate)
	}

	var rows []row
	err := query.Order("tokens DESC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var totalTokens int64
	for _, r := range rows {
		totalTokens += r.Tokens
	}

	limit := 5
	if len(rows) < limit {
		limit = len(rows)
	}

	result := make([]vo.ModelUsageItem, 0, limit+1)
	for i := 0; i < limit; i++ {
		var pct float64
		if totalTokens > 0 {
			pct = float64(rows[i].Tokens) / float64(totalTokens) * 100
		}
		result = append(result, vo.ModelUsageItem{
			ModelName:        rows[i].ModelName,
			TokenCount:       rows[i].Tokens,
			Percentage:       pct,
			PromptTokens:     rows[i].PromptTokens,
			CompletionTokens: rows[i].CompletionTokens,
		})
	}

	if len(rows) > limit {
		var otherTokens int64
		var otherPromptTokens int64
		var otherCompletionTokens int64
		for i := limit; i < len(rows); i++ {
			otherTokens += rows[i].Tokens
			otherPromptTokens += rows[i].PromptTokens
			otherCompletionTokens += rows[i].CompletionTokens
		}
		var pct float64
		if totalTokens > 0 {
			pct = float64(otherTokens) / float64(totalTokens) * 100
		}
		otherLabel := "其他"
		if config.Get().GetLanguage() == "en" {
			otherLabel = "Others"
		}
		result = append(result, vo.ModelUsageItem{
			ModelName:        otherLabel,
			TokenCount:       otherTokens,
			Percentage:       pct,
			PromptTokens:     otherPromptTokens,
			CompletionTokens: otherCompletionTokens,
		})
	}

	return result, nil
}

func parseDateRange(dateRange []string) (time.Time, time.Time) {
	startDate, _ := time.Parse("2006-01-02", dateRange[0])
	endDate, _ := time.Parse("2006-01-02", dateRange[len(dateRange)-1])
	return startDate, endDate.AddDate(0, 0, 1)
}

func (repo *DashboardRepository) GetUserTokenRank(dateRange []string) ([]vo.UserTokenRankItem, error) {
	type row struct {
		UserId     string
		Username   string
		TokenCount int64
	}

	query := repo.db().
		Table("user_daily_stats AS uds").
		Select("uds.user_id, COALESCE(u.username, uds.user_id) AS username, SUM(uds.prompt_tokens + uds.completion_tokens) AS tokenCount").
		Joins(fmt.Sprintf("LEFT JOIN %s AS u ON uds.user_id = u.user_id", repo.quoteUser())).
		Group("uds.user_id, u.username")

	if len(dateRange) > 0 {
		query = query.Where("uds.stat_date IN ?", dateRange)
	}

	var rows []row
	err := query.Order("tokenCount DESC").Limit(9).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var totalTokens int64
	for _, r := range rows {
		totalTokens += r.TokenCount
	}

	result := make([]vo.UserTokenRankItem, 0, len(rows))
	for _, r := range rows {
		var pct float64
		if totalTokens > 0 {
			pct = float64(r.TokenCount) / float64(totalTokens) * 100
		}
		result = append(result, vo.UserTokenRankItem{
			UserId:     r.UserId,
			Username:   r.Username,
			TokenCount: r.TokenCount,
			Percentage: pct,
		})
	}
	return result, nil
}

func (repo *DashboardRepository) GetActiveUserTrend(dateRange []string) ([]vo.ActiveUserTrendItem, error) {
	var rows []struct {
		Date  string
		Count int64
	}

	err := repo.db().
		Table("message").
		Select(fmt.Sprintf("%s AS date, COUNT(DISTINCT session.user_id) AS count", repo.dateExpr())).
		Joins("JOIN session ON message.session_id = session.session_id").
		Where(fmt.Sprintf("%s IN ?", repo.dateExpr()), dateRange).
		Group("date").
		Order("date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	dateMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		if len(r.Date) >= 10 {
			dateMap[r.Date[:10]] = r.Count
		}
	}

	result := make([]vo.ActiveUserTrendItem, 0, len(dateRange))
	for _, d := range dateRange {
		result = append(result, vo.ActiveUserTrendItem{
			Date:  d,
			Count: dateMap[d],
		})
	}
	return result, nil
}

func (repo *DashboardRepository) GetToolUsageRank(toolType string, dateRange []string) ([]vo.ToolUsageRankItem, error) {
	var rows []struct {
		ToolName string `gorm:"column:tool_name"`
		Count    int64  `gorm:"column:count"`
	}
	err := repo.db().
		Model(&po.ToolDailyStats{}).
		Select("tool_name, SUM(count) AS count").
		Where("tool_type = ? AND stat_date IN ?", toolType, dateRange).
		Group("tool_name").
		Order("count DESC").
		Limit(10).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]vo.ToolUsageRankItem, 0, len(rows))
	for _, r := range rows {
		result = append(result, vo.ToolUsageRankItem{
			Name:  r.ToolName,
			Count: r.Count,
		})
	}
	return result, nil
}

func (repo *DashboardRepository) GetUserTodayTokens(userId string) (int64, error) {
	today := time.Now().Format("2006-01-02")
	var total int64
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0)").
		Where("user_id = ? AND stat_date = ?", userId, today).
		Scan(&total).Error
	return total, err
}

func (repo *DashboardRepository) GetUserTotalTokens(userId string) (int64, error) {
	var total int64
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0)").
		Where("user_id = ?", userId).
		Scan(&total).Error
	return total, err
}

func (repo *DashboardRepository) GetUserTokensInDateRange(userId string, dateRange []string) (prompt int64, completion int64, err error) {
	type result struct {
		Prompt     int64
		Completion int64
	}
	var r result
	err = repo.db().
		Model(&po.UserDailyStats{}).
		Select("COALESCE(SUM(prompt_tokens), 0) AS prompt, COALESCE(SUM(completion_tokens), 0) AS completion").
		Where("user_id = ? AND stat_date IN ?", userId, dateRange).
		Scan(&r).Error
	return r.Prompt, r.Completion, err
}

func (repo *DashboardRepository) GetUserSparkline(userId string, dateRange []string) ([]vo.SparklineItem, error) {
	var rows []struct {
		StatDate string `gorm:"column:stat_date"`
		Tokens   int64  `gorm:"column:tokens"`
	}
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("stat_date, SUM(prompt_tokens + completion_tokens) AS tokens").
		Where("user_id = ? AND stat_date IN ?", userId, dateRange).
		Group("stat_date").
		Order("stat_date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]vo.SparklineItem, 0, len(dateRange))
	dateMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		dateMap[r.StatDate] = r.Tokens
	}
	for _, d := range dateRange {
		result = append(result, vo.SparklineItem{Date: d, Tokens: dateMap[d]})
	}
	return result, nil
}

func (repo *DashboardRepository) GetUserTokenTrend(userId string, dateRange []string) ([]vo.TokenTrendItem, error) {
	var rows []struct {
		StatDate         string `gorm:"column:stat_date"`
		PromptTokens     int64  `gorm:"column:prompt_tokens"`
		CompletionTokens int64  `gorm:"column:completion_tokens"`
	}
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("stat_date, COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens, COALESCE(SUM(completion_tokens), 0) AS completion_tokens").
		Where("user_id = ? AND stat_date IN ?", userId, dateRange).
		Group("stat_date").
		Order("stat_date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	dateMap := make(map[string]*vo.TokenTrendItem, len(rows))
	for i := range rows {
		ds := rows[i].StatDate
		dateMap[ds] = &vo.TokenTrendItem{
			Date:             ds,
			PromptTokens:     rows[i].PromptTokens,
			CompletionTokens: rows[i].CompletionTokens,
		}
	}

	result := make([]vo.TokenTrendItem, 0, len(dateRange))
	for _, d := range dateRange {
		if item, ok := dateMap[d]; ok {
			result = append(result, *item)
		} else {
			result = append(result, vo.TokenTrendItem{Date: d})
		}
	}
	return result, nil
}

func (repo *DashboardRepository) GetUserModelUsage(userId string, dateRange []string) ([]vo.ModelUsageItem, error) {
	deletedLabel := "已删除模型"
	if config.Get().GetLanguage() == "en" {
		deletedLabel = "Deleted Model"
	}

	type row struct {
		ModelName        string
		Tokens           int64
		PromptTokens     int64
		CompletionTokens int64
	}

	query := repo.db().
		Table("session").
		Select(fmt.Sprintf("CASE WHEN ai_model.ai_model_id IS NULL OR ai_model.deleted = 1 THEN '%s' ELSE ai_model.display_name END AS model_name, SUM(session.prompt_tokens_count + session.completion_tokens_count) AS tokens, SUM(session.prompt_tokens_count) AS prompt_tokens, SUM(session.completion_tokens_count) AS completion_tokens", deletedLabel)).
		Joins("LEFT JOIN ai_model ON session.ai_model_id = ai_model.ai_model_id").
		Where("session.deleted = 0 AND session.user_id = ?", userId).
		Group(fmt.Sprintf("CASE WHEN ai_model.ai_model_id IS NULL OR ai_model.deleted = 1 THEN '%s' ELSE ai_model.display_name END", deletedLabel))

	if len(dateRange) > 0 {
		startDate, endDate := parseDateRange(dateRange)
		query = query.Where("session.created >= ? AND session.created < ?", startDate, endDate)
	}

	var rows []row
	err := query.Order("tokens DESC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var totalTokens int64
	for _, r := range rows {
		totalTokens += r.Tokens
	}

	limit := 5
	if len(rows) < limit {
		limit = len(rows)
	}

	result := make([]vo.ModelUsageItem, 0, limit+1)
	for i := 0; i < limit; i++ {
		var pct float64
		if totalTokens > 0 {
			pct = float64(rows[i].Tokens) / float64(totalTokens) * 100
		}
		result = append(result, vo.ModelUsageItem{
			ModelName:        rows[i].ModelName,
			TokenCount:       rows[i].Tokens,
			Percentage:       pct,
			PromptTokens:     rows[i].PromptTokens,
			CompletionTokens: rows[i].CompletionTokens,
		})
	}

	if len(rows) > limit {
		var otherTokens int64
		var otherPromptTokens int64
		var otherCompletionTokens int64
		for i := limit; i < len(rows); i++ {
			otherTokens += rows[i].Tokens
			otherPromptTokens += rows[i].PromptTokens
			otherCompletionTokens += rows[i].CompletionTokens
		}
		var pct float64
		if totalTokens > 0 {
			pct = float64(otherTokens) / float64(totalTokens) * 100
		}
		otherLabel := "其他"
		if config.Get().GetLanguage() == "en" {
			otherLabel = "Others"
		}
		result = append(result, vo.ModelUsageItem{
			ModelName:        otherLabel,
			TokenCount:       otherTokens,
			Percentage:       pct,
			PromptTokens:     otherPromptTokens,
			CompletionTokens: otherCompletionTokens,
		})
	}

	return result, nil
}

func (repo *DashboardRepository) GetUserToolUsageRank(userId, toolType string, dateRange []string) ([]vo.ToolUsageRankItem, error) {
	var rows []struct {
		ToolName string `gorm:"column:tool_name"`
		Count    int64  `gorm:"column:count"`
	}
	err := repo.db().
		Model(&po.ToolDailyStats{}).
		Select("tool_name, SUM(count) AS count").
		Where("user_id = ? AND tool_type = ? AND stat_date IN ?", userId, toolType, dateRange).
		Group("tool_name").
		Order("count DESC").
		Limit(10).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]vo.ToolUsageRankItem, 0, len(rows))
	for _, r := range rows {
		result = append(result, vo.ToolUsageRankItem{
			Name:  r.ToolName,
			Count: r.Count,
		})
	}
	return result, nil
}

func (repo *DashboardRepository) GetUserSessionCounts(userId string) (total int64, newThisWeek int64, err error) {
	if err = repo.db().Model(&po.Session{}).Where("deleted = 0 AND user_id = ?", userId).Count(&total).Error; err != nil {
		return 0, 0, err
	}

	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	mondayStart := time.Date(now.Year(), now.Month(), now.Day()-int(weekday)+1, 0, 0, 0, 0, now.Location())

	if err = repo.db().Model(&po.Session{}).Where("deleted = 0 AND user_id = ? AND created >= ?", userId, mondayStart).Count(&newThisWeek).Error; err != nil {
		return 0, 0, err
	}
	return total, newThisWeek, nil
}

const dashboardCacheTTL = 10 * time.Minute

func (repo *DashboardRepository) GetDashboardCache(key string, result interface{}) bool {
	cached, err := repo.Redis().Get(context.Background(), key).Bytes()
	if err != nil {
		return false
	}
	if err := json.Unmarshal(cached, result); err != nil {
		return false
	}
	return true
}

func (repo *DashboardRepository) SetDashboardCache(key string, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	repo.Redis().Set(context.Background(), key, b, dashboardCacheTTL)
}

func (repo *DashboardRepository) InvalidateAllDashboardCache() {
	ctx := context.Background()
	const pattern = "dashboard:"
	if deleter, ok := repo.Redis().(infra.PatternDeleter); ok {
		deleter.DelByPattern(ctx, pattern)
		return
	}
	if client, ok := repo.Redis().(*redis.Client); ok {
		invalidateRedisByPattern(ctx, client, pattern)
	}
}

func invalidateRedisByPattern(ctx context.Context, client *redis.Client, pattern string) {
	var cursor uint64
	for {
		keys, next, err := client.Scan(ctx, cursor, pattern+"*", 100).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			client.Del(ctx, keys...)
		}
		if next == 0 {
			return
		}
		cursor = next
	}
}

func genDateRange(startDaysAgo, endDaysAgo int) []string {
	if startDaysAgo < endDaysAgo {
		return nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var result []string
	for i := startDaysAgo; i >= endDaysAgo; i-- {
		d := today.AddDate(0, 0, -i)
		result = append(result, d.Format("2006-01-02"))
	}
	return result
}

func (repo *DashboardRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
