package seed

import (
	"time"

	"raven/backend/vo"
)

// tokenTrendOffsets 定义 30 天 Token 趋势的固定偏移量与数值
// daysAgo 为距今天的天数（0=今天），prompt/completion 为固定值
// 刷新页面数据不变，但日期始终跟随当前时间
var tokenTrendOffsets = []struct {
	daysAgo          int
	promptTokens     int64
	completionTokens int64
}{
	// 早期（远→近，与数据库 ORDER BY statDate ASC 一致）
	{29, 15000, 6000},
	{28, 12000, 5000},
	{27, 17000, 7000},
	{26, 8000, 3000},
	{25, 22000, 9000},
	{24, 13000, 5000},
	{23, 9000, 3500},
	{22, 11000, 4500},
	{21, 30000, 12000},
	{20, 14000, 5500},
	{19, 10000, 4000},
	{18, 35000, 14000},
	{17, 16000, 6500},
	{16, 18000, 7500},
	{15, 12000, 5000},
	{14, 40000, 16000},
	{13, 15000, 6000},
	{12, 28000, 11000},
	{11, 18000, 7000},
	{10, 22000, 9000},
	{9, 20000, 8000},
	{8, 25000, 10000},
	{7, 32000, 13000},
	// 近一周
	{6, 28000, 11000},
	{5, 35000, 14000},
	{4, 48000, 20000},
	{3, 42000, 17000},
	{2, 38000, 15000},
	{1, 45000, 18000},
	{0, 52000, 22000},
}

// offsetDate 返回 daysAgo 天前的日期字符串 YYYY-MM-DD
func offsetDate(daysAgo int) string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return today.AddDate(0, 0, -daysAgo).Format("2006-01-02")
}

// BuildUserDashboard 构建用户仪表盘 mock 数据
// 日期基于当前时间动态生成，数值固定不变，刷新不跳动
func BuildUserDashboard() *vo.UserDashboardRsp {
	// ── Token 趋势（30 天）──
	tokenTrend := make([]vo.TokenTrendItem, 0, 30)
	for _, t := range tokenTrendOffsets {
		tokenTrend = append(tokenTrend, vo.TokenTrendItem{
			Date:             offsetDate(t.daysAgo),
			PromptTokens:     t.promptTokens,
			CompletionTokens: t.completionTokens,
		})
	}

	// ── Sparkline（近 7 天）──
	sparkline := make([]vo.SparklineItem, 0, 7)
	for _, t := range tokenTrendOffsets {
		if t.daysAgo > 6 {
			continue
		}
		sparkline = append(sparkline, vo.SparklineItem{
			Date:   offsetDate(t.daysAgo),
			Tokens: t.promptTokens + t.completionTokens,
		})
	}

	// ── Overview 聚合 ──
	var todayTokens, weekTokens, totalTokens int64
	for _, t := range tokenTrendOffsets {
		sum := t.promptTokens + t.completionTokens
		if t.daysAgo == 0 {
			todayTokens = sum
		}
		if t.daysAgo < 7 {
			weekTokens += sum
		}
		totalTokens += sum
	}

	// ── 工具使用排行（近 7 天汇总）──
	skillUsageRank := []vo.ToolUsageRankItem{
		{Name: "financial-report-analyzer", Count: 36},
		{Name: "data-viz-assistant", Count: 27},
		{Name: "lesson-planner", Count: 25},
		{Name: "essay-writer", Count: 18},
		{Name: "research-paper-helper", Count: 15},
		{Name: "tax-helper", Count: 12},
		{Name: "creative-copywriter", Count: 14},
		{Name: "budget-planner", Count: 7},
		{Name: "go-project-scaffold", Count: 8},
		{Name: "docker-compose-gen", Count: 8},
	}

	mcpUsageRank := []vo.ToolUsageRankItem{
		{Name: "filesystem", Count: 261},
		{Name: "postgres", Count: 128},
		{Name: "financial-data", Count: 81},
		{Name: "document-editor", Count: 45},
		{Name: "github", Count: 35},
		{Name: "arxiv-search", Count: 24},
		{Name: "brave-search", Count: 16},
		{Name: "translation", Count: 10},
		{Name: "weather", Count: 5},
		{Name: "email", Count: 10},
	}

	toolUsageRank := []vo.ToolUsageRankItem{
		{Name: "read_file", Count: 75},
		{Name: "write_file", Count: 44},
		{Name: "unique_id", Count: 14},
		{Name: "web_fetch", Count: 19},
		{Name: "file_url", Count: 7},
		{Name: "delete_file", Count: 5},
		{Name: "copy_file", Count: 3},
	}

	// ── 存储空间 ──
	storageStats := vo.UserStorageStats{
		UsedBytes:  2_150_000_000,  // ~2 GB
		FreeBytes:  47_850_000_000, // ~44.6 GB
		TotalBytes: 50_000_000_000, // ~46.6 GB
		Items: []vo.StorageUsageItem{
			{Name: "documents", BytesSize: 1_200_000_000, Percentage: 2.4},
			{Name: "projects", BytesSize: 680_000_000, Percentage: 1.36},
			{Name: "temp", BytesSize: 150_000_000, Percentage: 0.3},
			{Name: "images", BytesSize: 95_000_000, Percentage: 0.19},
			{Name: "videos", BytesSize: 20_000_000, Percentage: 0.04},
			{Name: "downloads", BytesSize: 5_000_000, Percentage: 0.01},
		},
	}

	return &vo.UserDashboardRsp{
		Overview: vo.UserDashboardOverview{
			TodayTokens:   todayTokens,
			WeekTokens:    weekTokens,
			TotalTokens:   totalTokens,
			TotalSessions: 12,
			NewSessions:   3,
			Sparkline:     sparkline,
		},
		SkillUsageRank: skillUsageRank,
		McpUsageRank:   mcpUsageRank,
		ToolUsageRank:  toolUsageRank,
		StorageStats:   storageStats,
	}
}

// ──────────────────────────────────────────────
// 管理员仪表盘 mock 数据
// ──────────────────────────────────────────────

// activeUserTrendOffsets 定义 30 天活跃用户趋势的固定偏移量与数值
// daysAgo 为距今天的天数，count 为当日去重活跃用户数
var activeUserTrendOffsets = []struct {
	daysAgo int
	count   int64
}{
	{29, 3},
	{28, 2},
	{27, 4},
	{26, 2},
	{25, 5},
	{24, 3},
	{23, 2},
	{22, 3},
	{21, 6},
	{20, 3},
	{19, 2},
	{18, 7},
	{17, 4},
	{16, 5},
	{15, 3},
	{14, 8},
	{13, 4},
	{12, 6},
	{11, 5},
	{10, 6},
	{9, 5},
	{8, 7},
	{7, 8},
	{6, 7},
	{5, 9},
	{4, 11},
	{3, 10},
	{2, 9},
	{1, 12},
	{0, 14},
}

// BuildAdminDashboard 构建管理员仪表盘 mock 数据
// 日期基于当前时间动态生成，数值固定不变，刷新不跳动
func BuildAdminDashboard() *vo.AdminDashboardRsp {
	// ── Token 趋势（30 天）──
	tokenTrend := make([]vo.TokenTrendItem, 0, 30)
	for _, t := range tokenTrendOffsets {
		tokenTrend = append(tokenTrend, vo.TokenTrendItem{
			Date:             offsetDate(t.daysAgo),
			PromptTokens:     t.promptTokens,
			CompletionTokens: t.completionTokens,
		})
	}

	// ── Sparkline（近 7 天）──
	sparkline := make([]vo.SparklineItem, 0, 7)
	for _, t := range tokenTrendOffsets {
		if t.daysAgo > 6 {
			continue
		}
		sparkline = append(sparkline, vo.SparklineItem{
			Date:   offsetDate(t.daysAgo),
			Tokens: t.promptTokens + t.completionTokens,
		})
	}

	// ── Overview 聚合 ──
	var todayTokens, weekTokens int64
	var thisWeekActiveUsers, lastWeekActiveUsers int
	for _, t := range tokenTrendOffsets {
		sum := t.promptTokens + t.completionTokens
		if t.daysAgo == 0 {
			todayTokens = sum
		}
		if t.daysAgo < 7 {
			weekTokens += sum
		}
	}

	for _, t := range activeUserTrendOffsets {
		if t.daysAgo < 7 {
			thisWeekActiveUsers += int(t.count)
		} else if t.daysAgo >= 7 && t.daysAgo < 14 {
			lastWeekActiveUsers += int(t.count)
		}
	}

	// 环比变化率
	var activeUsersDiff float64
	if lastWeekActiveUsers > 0 {
		activeUsersDiff = float64(thisWeekActiveUsers-lastWeekActiveUsers) / float64(lastWeekActiveUsers) * 100
	}

	// ── 工具使用排行（近 7 天汇总）──
	skillUsageRank := []vo.ToolUsageRankItem{
		{Name: "financial-report-analyzer", Count: 142},
		{Name: "data-viz-assistant", Count: 97},
		{Name: "lesson-planner", Count: 85},
		{Name: "essay-writer", Count: 63},
		{Name: "research-paper-helper", Count: 51},
		{Name: "tax-helper", Count: 44},
		{Name: "creative-copywriter", Count: 38},
		{Name: "budget-planner", Count: 29},
		{Name: "go-project-scaffold", Count: 22},
		{Name: "docker-compose-gen", Count: 18},
	}

	mcpUsageRank := []vo.ToolUsageRankItem{
		{Name: "filesystem", Count: 890},
		{Name: "postgres", Count: 456},
		{Name: "financial-data", Count: 287},
		{Name: "document-editor", Count: 163},
		{Name: "github", Count: 124},
		{Name: "arxiv-search", Count: 89},
		{Name: "brave-search", Count: 57},
		{Name: "translation", Count: 34},
		{Name: "weather", Count: 18},
		{Name: "email", Count: 29},
	}

	toolUsageRank := []vo.ToolUsageRankItem{
		{Name: "read_file", Count: 312},
		{Name: "write_file", Count: 187},
		{Name: "web_fetch", Count: 76},
		{Name: "unique_id", Count: 52},
		{Name: "file_url", Count: 31},
		{Name: "delete_file", Count: 19},
		{Name: "copy_file", Count: 11},
	}

	return &vo.AdminDashboardRsp{
		Overview: vo.AdminDashboardOverview{
			ActiveUsers:     thisWeekActiveUsers,
			ActiveUsersDiff: activeUsersDiff,
			TotalSessions:   287,
			NewSessions:     42,
			WeekTokens:      weekTokens,
			TodayTokens:     todayTokens,
			EnabledModels:   6,
			Sparkline:       sparkline,
		},
		SkillUsageRank: skillUsageRank,
		McpUsageRank:   mcpUsageRank,
		ToolUsageRank:  toolUsageRank,
	}
}

// BuildAdminTokenTrend 构建管理员 Token 趋势 mock 数据
// days 参数控制返回的天数范围，日期基于当前时间
func BuildAdminTokenTrend(days int) *vo.TokenTrendRsp {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	// 30 天以内的数据直接从 tokenTrendOffsets 截取
	if days <= len(tokenTrendOffsets) {
		items := make([]vo.TokenTrendItem, 0, days)
		for _, t := range tokenTrendOffsets {
			if t.daysAgo >= days {
				continue
			}
			items = append(items, vo.TokenTrendItem{
				Date:             offsetDate(t.daysAgo),
				PromptTokens:     t.promptTokens,
				CompletionTokens: t.completionTokens,
			})
		}
		return &vo.TokenTrendRsp{Items: items}
	}

	// 超过 30 天时，先用早期低活跃数据填充 31~90 天，再追加 30 天数据
	// 顺序：最旧在前（daysAgo 大）→ 最新在后（daysAgo 小），保证日期升序
	earlyPattern := []struct {
		prompt, completion int64
	}{
		{25000, 10000},
		{18000, 7000},
		{22000, 9000},
		{10000, 4000},
		{30000, 12000},
		{15000, 6000},
		{35000, 14000},
	}

	items := make([]vo.TokenTrendItem, 0, days)
	// 早期：从 daysAgo=days-1（最旧）到 daysAgo=30
	for i := days - 1; i >= len(tokenTrendOffsets); i-- {
		p := earlyPattern[i%len(earlyPattern)]
		items = append(items, vo.TokenTrendItem{
			Date:             offsetDate(i),
			PromptTokens:     p.prompt,
			CompletionTokens: p.completion,
		})
	}
	// 近 30 天：daysAgo 29→0（tokenTrendOffsets 本身已是旧→新顺序）
	for _, t := range tokenTrendOffsets {
		items = append(items, vo.TokenTrendItem{
			Date:             offsetDate(t.daysAgo),
			PromptTokens:     t.promptTokens,
			CompletionTokens: t.completionTokens,
		})
	}

	return &vo.TokenTrendRsp{Items: items}
}

// BuildAdminActiveUserTrend 构建管理员活跃用户趋势 mock 数据
// days 参数控制返回的天数范围，日期基于当前时间
func BuildAdminActiveUserTrend(days int) *vo.ActiveUserTrendRsp {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	// 30 天以内的数据直接从 activeUserTrendOffsets 截取
	if days <= len(activeUserTrendOffsets) {
		items := make([]vo.ActiveUserTrendItem, 0, days)
		for _, t := range activeUserTrendOffsets {
			if t.daysAgo >= days {
				continue
			}
			items = append(items, vo.ActiveUserTrendItem{
				Date:  offsetDate(t.daysAgo),
				Count: t.count,
			})
		}
		return &vo.ActiveUserTrendRsp{Items: items}
	}

	// 超过 30 天时，先用早期低活跃数据填充 31~90 天，再追加 30 天数据
	// 顺序：最旧在前（daysAgo 大）→ 最新在后（daysAgo 小），保证日期升序
	earlyPattern := []int64{2, 1, 3, 1, 2, 2, 1}

	items := make([]vo.ActiveUserTrendItem, 0, days)
	// 早期：从 daysAgo=days-1（最旧）到 daysAgo=30
	for i := days - 1; i >= len(activeUserTrendOffsets); i-- {
		items = append(items, vo.ActiveUserTrendItem{
			Date:  offsetDate(i),
			Count: earlyPattern[i%len(earlyPattern)],
		})
	}
	// 近 30 天：daysAgo 29→0（activeUserTrendOffsets 本身已是旧→新顺序）
	for _, t := range activeUserTrendOffsets {
		items = append(items, vo.ActiveUserTrendItem{
			Date:  offsetDate(t.daysAgo),
			Count: t.count,
		})
	}

	return &vo.ActiveUserTrendRsp{Items: items}
}

// BuildUserTokenTrend 构建用户 Token 趋势 mock 数据
// days 参数控制返回的天数范围，日期基于当前时间
func BuildUserTokenTrend(days int) *vo.TokenTrendRsp {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	// 30 天以内的数据直接从 tokenTrendOffsets 截取
	if days <= len(tokenTrendOffsets) {
		items := make([]vo.TokenTrendItem, 0, days)
		for _, t := range tokenTrendOffsets {
			if t.daysAgo >= days {
				continue
			}
			items = append(items, vo.TokenTrendItem{
				Date:             offsetDate(t.daysAgo),
				PromptTokens:     t.promptTokens,
				CompletionTokens: t.completionTokens,
			})
		}
		return &vo.TokenTrendRsp{Items: items}
	}

	// 超过 30 天时，先用早期低活跃数据填充 31~90 天，再追加 30 天数据
	// 顺序：最旧在前（daysAgo 大）→ 最新在后（daysAgo 小），保证日期升序
	earlyPattern := []struct {
		prompt, completion int64
	}{
		{8000, 3000},
		{5000, 2000},
		{6000, 2500},
		{3000, 1200},
		{7000, 2800},
		{4000, 1500},
		{9000, 3500},
	}

	items := make([]vo.TokenTrendItem, 0, days)
	// 早期：从 daysAgo=days-1（最旧）到 daysAgo=30
	for i := days - 1; i >= len(tokenTrendOffsets); i-- {
		p := earlyPattern[i%len(earlyPattern)]
		items = append(items, vo.TokenTrendItem{
			Date:             offsetDate(i),
			PromptTokens:     p.prompt,
			CompletionTokens: p.completion,
		})
	}
	// 近 30 天：daysAgo 29→0（tokenTrendOffsets 本身已是旧→新顺序）
	for _, t := range tokenTrendOffsets {
		items = append(items, vo.TokenTrendItem{
			Date:             offsetDate(t.daysAgo),
			PromptTokens:     t.promptTokens,
			CompletionTokens: t.completionTokens,
		})
	}

	return &vo.TokenTrendRsp{Items: items}
}

// BuildAdminModelUsage 构建管理员模型使用分布 Mock 数据
// days 参数仅影响返回的 items 数量（以模拟时间切片效果），数据内容与日期无关
func BuildAdminModelUsage(days int) *vo.ModelUsageRsp {
	items := []vo.ModelUsageItem{
		{ModelName: "deepseek-v4-flash", TokenCount: 385000, Percentage: 39.9, PromptTokens: 220000, CompletionTokens: 165000},
		{ModelName: "MiniMax-M2.7", TokenCount: 248000, Percentage: 25.7, PromptTokens: 140000, CompletionTokens: 108000},
		{ModelName: "deepseek-v4-pro", TokenCount: 172000, Percentage: 17.8, PromptTokens: 100000, CompletionTokens: 72000},
		{ModelName: "GPT-5.5", TokenCount: 96000, Percentage: 9.9, PromptTokens: 55000, CompletionTokens: 41000},
		{ModelName: "Others", TokenCount: 64000, Percentage: 6.7, PromptTokens: 38000, CompletionTokens: 26000},
	}
	if days <= 7 {
		items = items[:min(len(items), 3)]
	}
	return &vo.ModelUsageRsp{Items: items}
}

// BuildUserModelUsage 构建用户模型使用分布 Mock 数据
// days 参数仅影响返回的 items 数量（以模拟时间切片效果），数据内容与日期无关
func BuildUserModelUsage(days int) *vo.ModelUsageRsp {
	items := []vo.ModelUsageItem{
		{ModelName: "deepseek-v4-flash", TokenCount: 1250000, Percentage: 32.1, PromptTokens: 700000, CompletionTokens: 550000},
		{ModelName: "MiniMax-M2.7", TokenCount: 820000, Percentage: 21.0, PromptTokens: 450000, CompletionTokens: 370000},
		{ModelName: "deepseek-v4-pro", TokenCount: 610000, Percentage: 15.7, PromptTokens: 350000, CompletionTokens: 260000},
		{ModelName: "GPT-5.5", TokenCount: 480000, Percentage: 12.3, PromptTokens: 280000, CompletionTokens: 200000},
		{ModelName: "Claude-Sonnet-4.6", TokenCount: 350000, Percentage: 9.0, PromptTokens: 200000, CompletionTokens: 150000},
		{ModelName: "Qwen-Max", TokenCount: 220000, Percentage: 5.6, PromptTokens: 130000, CompletionTokens: 90000},
		{ModelName: "Others", TokenCount: 170000, Percentage: 4.3, PromptTokens: 100000, CompletionTokens: 70000},
	}
	if days <= 7 {
		items = items[:min(len(items), 4)]
	}
	return &vo.ModelUsageRsp{Items: items}
}

// BuildAdminUserTokenRank 构建管理员用户 Token 消耗排行 Mock 数据
// days 参数仅影响返回的 items 数量（以模拟时间切片效果），数据内容与日期无关
func BuildAdminUserTokenRank(days int) *vo.UserTokenRankRsp {
	items := []vo.UserTokenRankItem{
		{UserId: "90a431bee756432492c134f510bad949", Username: "demo_admin", TokenCount: 965000, Percentage: 24.8},
		{UserId: "a1b2c3d4e5f6", Username: "zhang_wei", TokenCount: 640000, Percentage: 16.4},
		{UserId: "g7h8i9j0k1l2", Username: "li_na", TokenCount: 520000, Percentage: 13.4},
		{UserId: "m3n4o5p6q7r8", Username: "wang_fang", TokenCount: 410000, Percentage: 10.5},
		{UserId: "s9t0u1v2w3x4", Username: "chen_ming", TokenCount: 340000, Percentage: 8.7},
		{UserId: "y5z6a7b8c9d0", Username: "liu_yang", TokenCount: 280000, Percentage: 7.2},
		{UserId: "e1f2g3h4i5j6", Username: "zhao_jun", TokenCount: 220000, Percentage: 5.6},
		{UserId: "k7l8m9n0o1p2", Username: "sun_li", TokenCount: 180000, Percentage: 4.6},
		{UserId: "q3r4s5t6u7v8", Username: "zhou_hua", TokenCount: 150000, Percentage: 3.9},
	}
	if days <= 7 {
		items = items[:min(len(items), 5)]
	}
	return &vo.UserTokenRankRsp{Items: items}
}
