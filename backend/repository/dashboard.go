package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/config"

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

// DashboardRepository 仪表盘数据仓库
// 提供 Token 趋势、活跃趋势、模型/用户/工具排行等聚合查询
type DashboardRepository struct {
	freedom.Repository
}

// quoteUser 返回当前数据库正确引号包裹的 user 表名
// PG/SQLite 用双引号，MySQL 用反引号（user 是 PG 保留字）
func (repo *DashboardRepository) quoteUser() string {
	if config.Get().System.DBType == "mysql" {
		return "`user`"
	}
	return `"user"`
}
// dateExpr 根据数据库类型生成从毫秒时间戳提取日期的 SQL 表达式
// message.timestamp 为 Unix 毫秒，需转换为 YYYY-MM-DD 格式
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

// GetActiveUsersInRange 查询指定天数范围内的活跃用户数（从 message JOIN session 去重 userId）
// message 表通过 sessionId 关联 session 表获取 userId
// startDaysAgo 为起始偏移天数（含），endDaysAgo 为结束偏移天数（不含），均为正数
// 示例：GetActiveUsersInRange(7, 0) 返回近 7 日活跃用户数
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

// GetSessionCounts 查询会话总数和本周新增会话数
func (repo *DashboardRepository) GetSessionCounts() (total int64, newThisWeek int64, err error) {
	if err = repo.db().Model(&po.Session{}).Where("deleted = 0").Count(&total).Error; err != nil {
		return 0, 0, err
	}

	// 本周一 00:00:00
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

// GetTodayTokens 查询今日 Token 消耗（从 user_daily_stats）
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

// GetEnabledModelCount 查询启用模型数
func (repo *DashboardRepository) GetEnabledModelCount() (int64, error) {
	var count int64
	err := repo.db().Model(&po.AIModel{}).Where("status = 1 AND deleted = 0").Count(&count).Error
	return count, err
}

// GetTokensInDateRange 查询 user_daily_stats 中指定日期范围的 Token 总和
// 返回 promptTokens 总和和 completionTokens 总和
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

// GetSparkline 查询近 7 日 Token 消耗迷你趋势（按天聚合）
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

// GetTokenTrend 查询 Token 消耗趋势（按天聚合 prompt + completion）
func (repo *DashboardRepository) GetTokenTrend(dateRange []string) ([]vo.TokenTrendItem, error) {
	var rows []struct {
		StatDate           string `gorm:"column:stat_date"`
		PromptTokens       int64  `gorm:"column:prompt_tokens"`
		PromptCachedTokens int64  `gorm:"column:prompt_cached_tokens"`
		CompletionTokens   int64  `gorm:"column:completion_tokens"`
	}
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("stat_date, COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens, COALESCE(SUM(prompt_cached_tokens), 0) AS prompt_cached_tokens, COALESCE(SUM(completion_tokens), 0) AS completion_tokens").
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
			Date:               ds,
			PromptTokens:       rows[i].PromptTokens,
			PromptCachedTokens: rows[i].PromptCachedTokens,
			CompletionTokens:   rows[i].CompletionTokens,
		}
	}

	// 保证返回完整的日期序列，缺失日期填 0
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

// GetModelUsage 查询模型使用分布
// 按模型显示名聚合 Token，JOIN ai_model 获取模型名称。
// 已删除模型（ai_model.deleted=1 或缺失）的 session 会合并为「已删除模型」单条，
// 避免按 aiModelId 分组产生多条 Unknown 重复行污染 Top 5 与百分比。
// dateRange 可选：传入非空则按 session.created 过滤日期范围；传入 nil/空则不过滤（全量）。
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

	// 计算总和，取 Top 5，其余合并为「其他」
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

// parseDateRange 将日期字符串列表转换为 time.Time 范围
// dateRange[0] 为开始日期，dateRange[len-1] 为结束日期，endDate 为结束日期 + 1 天（开区间）
func parseDateRange(dateRange []string) (time.Time, time.Time) {
	startDate, _ := time.Parse("2006-01-02", dateRange[0])
	endDate, _ := time.Parse("2006-01-02", dateRange[len(dateRange)-1])
	return startDate, endDate.AddDate(0, 0, 1)
}

// GetUserTokenRank 查询用户 Token 消耗排行（Top 9）
// 从 user_daily_stats 聚合，JOIN user 获取用户名
// dateRange 可选：传入非空则按 stat_date 过滤日期范围；传入 nil/空则不过滤（全量）
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

// GetActiveUserTrend 查询活跃用户趋势（从 message JOIN session 按天去重 userId）
// 返回完整的日期序列，没有数据的日期填 0
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
		// dateExpr() 在不同数据库返回不同类型：SQLite 返回 "2006-01-02" 纯文本，
		// MySQL/PG 的 DATE 函数返回值被 driver 扫描为 time.Time → GORM 转成 RFC3339 字符串。
		// 统一截取前 10 字符（YYYY-MM-DD）确保与 dateRange 的 key 精确匹配。
		if len(r.Date) >= 10 {
			dateMap[r.Date[:10]] = r.Count
		}
	}

	// 保证返回完整的日期序列，缺失日期填 0
	result := make([]vo.ActiveUserTrendItem, 0, len(dateRange))
	for _, d := range dateRange {
		result = append(result, vo.ActiveUserTrendItem{
			Date:  d,
			Count: dateMap[d],
		})
	}
	return result, nil
}

// GetToolUsageRank 查询指定类型的工具使用排行（近 7 天 Top 10）
// toolType: skill / mcp / tool
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

// ────────────────────────────────────────────────────
// 用户范围查询（按 userId 过滤）
// ────────────────────────────────────────────────────

// GetUserTodayTokens 查询指定用户今日 Token 消耗（从 user_daily_stats）
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

// GetUserDailyTokenLimit 查询指定用户的每日 Token 限额（单位 M，0=不限制）
func (repo *DashboardRepository) GetUserDailyTokenLimit(userId string) (int, error) {
	var limit int
	err := repo.db().
		Model(&po.User{}).
		Select("daily_token_limit").
		Where("user_id = ? AND deleted = 0", userId).
		Scan(&limit).Error
	return limit, err
}

// GetUserTotalTokens 查询指定用户历史累计 Token 消耗（从 user_daily_stats 全量聚合）
func (repo *DashboardRepository) GetUserTotalTokens(userId string) (int64, error) {
	var total int64
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0)").
		Where("user_id = ?", userId).
		Scan(&total).Error
	return total, err
}

// GetUserTokensInDateRange 查询指定用户在日期范围内的 Token 总和
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

// GetUserSparkline 查询指定用户近 7 日 Token 消耗迷你趋势（按天聚合）
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

// GetUserTokenTrend 查询指定用户 Token 消耗趋势（按天聚合 prompt + completion）
func (repo *DashboardRepository) GetUserTokenTrend(userId string, dateRange []string) ([]vo.TokenTrendItem, error) {
	var rows []struct {
		StatDate           string `gorm:"column:stat_date"`
		PromptTokens       int64  `gorm:"column:prompt_tokens"`
		PromptCachedTokens int64  `gorm:"column:prompt_cached_tokens"`
		CompletionTokens   int64  `gorm:"column:completion_tokens"`
	}
	err := repo.db().
		Model(&po.UserDailyStats{}).
		Select("stat_date, COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens, COALESCE(SUM(prompt_cached_tokens), 0) AS prompt_cached_tokens, COALESCE(SUM(completion_tokens), 0) AS completion_tokens").
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
			Date:               ds,
			PromptTokens:       rows[i].PromptTokens,
			PromptCachedTokens: rows[i].PromptCachedTokens,
			CompletionTokens:   rows[i].CompletionTokens,
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

// GetUserModelUsage 查询指定用户的模型使用分布
// 按模型显示名聚合 Token，JOIN ai_model 获取模型名称。
// 已删除模型（ai_model.deleted=1 或缺失）的 session 会合并为「已删除模型」单条。
// dateRange 可选：传入非空则按 session.created 过滤日期范围；传入 nil/空则不过滤（全量）。
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

// GetUserToolUsageRank 查询指定用户指定类型的工具使用排行（近 7 天 Top 10）
// toolType: skill / mcp / tool
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

// GetUserSessionCounts 查询指定用户的会话总数和本周新增会话数
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

// ═══════════════════════════════════════════════════════════════
// 缓存
// ═══════════════════════════════════════════════════════════════

const dashboardCacheTTL = 10 * time.Minute

// getDashboardCache 从 Redis 读取缓存数据，反序列化到 result
// 缓存不存在或解析失败返回 false
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

// setDashboardCache 将数据序列化后写入 Redis，TTL 为 10 分钟
func (repo *DashboardRepository) SetDashboardCache(key string, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	repo.Redis().Set(context.Background(), key, b, dashboardCacheTTL)
}

// InvalidateAllDashboardCache 失效所有仪表盘相关缓存（管理员 + 所有用户）
// 模型增删改、启停状态变化等会影响多个用户的仪表盘聚合，调用此方法可立即让前端看到最新数据
// 适配本地 go-cache（infra.PatternDeleter）与真实 Redis（SCAN+DEL）两种后端
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

// invalidateRedisByPattern 真实 Redis 后端：使用 SCAN 迭代匹配前缀并删除
// 避免 KEYS 在大 keyspace 下阻塞 Redis
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

// genDateRange 生成日期范围字符串列表（YYYY-MM-DD）
// startDaysAgo 为起始偏移天数（含，如 6 表示 6 天前），endDaysAgo 为结束偏移天数（含，如 0 表示今天）
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

// db 获取数据库连接
func (repo *DashboardRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
