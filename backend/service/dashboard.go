package service

import (
	"fmt"

	"goraven/backend/infra"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/core/sandbox"
	"goraven/util"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *DashboardService {
			return &DashboardService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *DashboardService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// DashboardService 仪表盘服务
// 提供管理员全局仪表盘和用户个人仪表盘的数据聚合
type DashboardService struct {
	Worker        freedom.Worker
	DashboardRepo *repository.DashboardRepository
}

// ═══════════════════════════════════════════════════════════════
// 管理员仪表盘（全局范围）
// ═══════════════════════════════════════════════════════════════

// GetAdminDashboard 获取管理员仪表盘完整聚合数据
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetAdminDashboard() (*vo.AdminDashboardRsp, error) {
	const cacheKey = "dashboard:admin"

	rsp := &vo.AdminDashboardRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	last7 := util.GenLastNDates(7)      // 近 7 天（含今天）
	thisWeek := util.GenThisWeekDates() // 本周一至周日

	// ── 系统脉搏 ──
	overview, err := service.buildAdminOverview(last7, thisWeek)
	if err != nil {
		return nil, err
	}
	rsp.Overview = *overview

	// ── 近一周技能/MCP/工具使用排行 ──
	skillRank, err := service.DashboardRepo.GetToolUsageRank("skill", last7)
	if err != nil {
		return nil, err
	}
	rsp.SkillUsageRank = skillRank

	mcpRank, err := service.DashboardRepo.GetToolUsageRank("mcp", last7)
	if err != nil {
		return nil, err
	}
	rsp.McpUsageRank = mcpRank

	toolRank, err := service.DashboardRepo.GetToolUsageRank("tool", last7)
	if err != nil {
		return nil, err
	}
	rsp.ToolUsageRank = toolRank

	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

// GetAdminTokenTrend 获取管理员全局 Token 消耗趋势（支持切换时间粒度）
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetAdminTokenTrend(days int) (*vo.TokenTrendRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:admin:tokenTrend:%d", days)

	rsp := &vo.TokenTrendRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	dateRange := util.GenLastNDates(days)
	items, err := service.DashboardRepo.GetTokenTrend(dateRange)
	if err != nil {
		return nil, err
	}
	rsp = &vo.TokenTrendRsp{Items: items}
	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

// GetAdminModelUsage 获取管理员全局模型使用分布（支持切换时间粒度）
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetAdminModelUsage(days int) (*vo.ModelUsageRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:admin:modelUsage:%d", days)

	rsp := &vo.ModelUsageRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	dateRange := util.GenLastNDates(days)
	items, err := service.DashboardRepo.GetModelUsage(dateRange)
	if err != nil {
		return nil, err
	}
	rsp = &vo.ModelUsageRsp{Items: items}
	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

// GetAdminUserTokenRank 获取管理员用户 Token 消耗排行（支持切换时间粒度）
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetAdminUserTokenRank(days int) (*vo.UserTokenRankRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:admin:userTokenRank:%d", days)

	rsp := &vo.UserTokenRankRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	dateRange := util.GenLastNDates(days)
	items, err := service.DashboardRepo.GetUserTokenRank(dateRange)
	if err != nil {
		return nil, err
	}
	rsp = &vo.UserTokenRankRsp{Items: items}
	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

// GetAdminActiveUserTrend 获取管理员全局活跃用户趋势（支持切换时间粒度）
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetAdminActiveUserTrend(days int) (*vo.ActiveUserTrendRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:admin:activeUsers:%d", days)

	rsp := &vo.ActiveUserTrendRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	dateRange := util.GenLastNDates(days)
	items, err := service.DashboardRepo.GetActiveUserTrend(dateRange)
	if err != nil {
		return nil, err
	}
	rsp = &vo.ActiveUserTrendRsp{Items: items}
	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

// buildAdminOverview 构建管理员系统脉搏核心指标
func (service *DashboardService) buildAdminOverview(last7, thisWeek []string) (*vo.AdminDashboardOverview, error) {
	overview := &vo.AdminDashboardOverview{}

	// 近 7 日活跃用户
	activeUsers, err := service.DashboardRepo.GetActiveUsersInRange(7, 0)
	if err != nil {
		return nil, err
	}
	overview.ActiveUsers = int(activeUsers)

	// 环比上周
	prevActive, err := service.DashboardRepo.GetActiveUsersInRange(14, 7)
	if err != nil {
		return nil, err
	}
	if prevActive > 0 {
		overview.ActiveUsersDiff = float64(activeUsers-prevActive) / float64(prevActive) * 100
	}

	// 会话总数 + 本周新增
	totalSessions, newThisWeek, err := service.DashboardRepo.GetSessionCounts()
	if err != nil {
		return nil, err
	}
	overview.TotalSessions = totalSessions
	overview.NewSessions = newThisWeek

	// 本周 + 今日 Token 消耗
	prompt, completion, err := service.DashboardRepo.GetTokensInDateRange(thisWeek)
	if err != nil {
		return nil, err
	}
	overview.WeekTokens = prompt + completion

	todayTokens, err := service.DashboardRepo.GetTodayTokens()
	if err != nil {
		return nil, err
	}
	overview.TodayTokens = todayTokens

	// 启用模型数
	modelCount, err := service.DashboardRepo.GetEnabledModelCount()
	if err != nil {
		return nil, err
	}
	overview.EnabledModels = modelCount

	// 7 日迷你趋势
	sparkline, err := service.DashboardRepo.GetSparkline(last7)
	if err != nil {
		return nil, err
	}
	overview.Sparkline = sparkline

	return overview, nil
}

// ═══════════════════════════════════════════════════════════════
// 用户仪表盘（个人范围，按 userId 过滤）
// ═══════════════════════════════════════════════════════════════

// GetUserDashboard 获取用户个人仪表盘完整聚合数据
// 一次返回脉搏指标、工具排行、存储统计
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetUserDashboard(userId string) (*vo.UserDashboardRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:user:%s", userId)

	rsp := &vo.UserDashboardRsp{}
	cacheHit := service.DashboardRepo.GetDashboardCache(cacheKey, rsp)

	if !cacheHit {
		last7 := util.GenLastNDates(7)      // 近 7 天（含今天）
		thisWeek := util.GenThisWeekDates() // 本周一至周日

		// ── 个人脉搏 ──
		overview, err := service.buildUserOverview(userId, last7, thisWeek)
		if err != nil {
			return nil, err
		}
		rsp.Overview = *overview

		// ── 近一周技能/MCP/工具使用排行 ──
		skillRank, err := service.DashboardRepo.GetUserToolUsageRank(userId, "skill", last7)
		if err != nil {
			return nil, err
		}
		rsp.SkillUsageRank = skillRank

		mcpRank, err := service.DashboardRepo.GetUserToolUsageRank(userId, "mcp", last7)
		if err != nil {
			return nil, err
		}
		rsp.McpUsageRank = mcpRank

		toolRank, err := service.DashboardRepo.GetUserToolUsageRank(userId, "tool", last7)
		if err != nil {
			return nil, err
		}
		rsp.ToolUsageRank = toolRank

		// ── 存储空间 ──
		storageStats, err := service.buildStorageStats(userId)
		if err != nil {
			return nil, err
		}
		rsp.StorageStats = storageStats

		service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	}

	// ── 日限额 / 日使用 / 今日：实时查询，不走缓存 ──
	dailyLimit, err := service.DashboardRepo.GetUserDailyTokenLimit(userId)
	if err != nil {
		return nil, err
	}
	todayUsed, err := service.DashboardRepo.GetUserTodayTokens(userId)
	if err != nil {
		return nil, err
	}
	rsp.Overview.DailyTokenLimit = dailyLimit
	rsp.Overview.TodayTokens = todayUsed

	return rsp, nil
}

// GetUserTokenTrend 获取用户个人 Token 消耗趋势（支持切换时间粒度）
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetUserTokenTrend(userId string, days int) (*vo.TokenTrendRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:user:%s:tokenTrend:%d", userId, days)

	rsp := &vo.TokenTrendRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	dateRange := util.GenLastNDates(days)
	items, err := service.DashboardRepo.GetUserTokenTrend(userId, dateRange)
	if err != nil {
		return nil, err
	}
	rsp = &vo.TokenTrendRsp{Items: items}
	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

// GetUserModelUsage 获取用户个人模型使用分布（支持切换时间粒度）
// 优先读缓存（TTL 10 分钟），未命中则查库并写缓存
func (service *DashboardService) GetUserModelUsage(userId string, days int) (*vo.ModelUsageRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:user:%s:modelUsage:%d", userId, days)

	rsp := &vo.ModelUsageRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	dateRange := util.GenLastNDates(days)
	items, err := service.DashboardRepo.GetUserModelUsage(userId, dateRange)
	if err != nil {
		return nil, err
	}
	rsp = &vo.ModelUsageRsp{Items: items}
	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

// buildUserOverview 构建用户个人脉搏核心指标
func (service *DashboardService) buildUserOverview(userId string, last7, thisWeek []string) (*vo.UserDashboardOverview, error) {
	overview := &vo.UserDashboardOverview{}

	// 会话总数 + 本周新增
	totalSessions, newThisWeek, err := service.DashboardRepo.GetUserSessionCounts(userId)
	if err != nil {
		return nil, err
	}
	overview.TotalSessions = totalSessions
	overview.NewSessions = newThisWeek

	// 本周 Token 消耗
	prompt, completion, err := service.DashboardRepo.GetUserTokensInDateRange(userId, thisWeek)
	if err != nil {
		return nil, err
	}
	overview.WeekTokens = prompt + completion

	// 今日 Token 消耗
	todayTokens, err := service.DashboardRepo.GetUserTodayTokens(userId)
	if err != nil {
		return nil, err
	}
	overview.TodayTokens = todayTokens

	// 历史累计 Token 消耗
	totalTokens, err := service.DashboardRepo.GetUserTotalTokens(userId)
	if err != nil {
		return nil, err
	}
	overview.TotalTokens = totalTokens

	// 7 日迷你趋势
	sparkline, err := service.DashboardRepo.GetUserSparkline(userId, last7)
	if err != nil {
		return nil, err
	}
	overview.Sparkline = sparkline

	return overview, nil
}

// buildStorageStats 通过 Sandbox 接口获取用户空间各子目录的存储使用量，计算占比
// skills 目录为系统技能安装目录，不在仪表盘展示
// totalBytes 为磁盘总容量，freeBytes 为磁盘剩余可用空间，pct 分母为磁盘总容量
func (service *DashboardService) buildStorageStats(userId string) (vo.UserStorageStats, error) {
	box, boxerr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if boxerr != nil {
		return vo.UserStorageStats{}, boxerr
	}
	usage, err := box.GetStorageUsage()
	if err != nil {
		return vo.UserStorageStats{}, err
	}

	capacity, err := box.GetStorageCapacity()
	if err != nil {
		return vo.UserStorageStats{}, err
	}

	var usedBytes int64
	for _, u := range usage {
		if u.Name == "skills" {
			continue
		}
		usedBytes += u.BytesSize
	}

	totalBytes := capacity.TotalBytes
	freeBytes := capacity.FreeBytes
	if totalBytes > 10*1024*1024*1024*1024 {
		totalBytes = 0
		freeBytes = 0
	}

	items := make([]vo.StorageUsageItem, 0, len(usage))
	for _, u := range usage {
		if u.Name == "skills" {
			continue
		}
		var pct float64
		if totalBytes > 0 {
			pct = float64(u.BytesSize) / float64(totalBytes) * 100
		}
		items = append(items, vo.StorageUsageItem{
			Name:       u.Name,
			BytesSize:  u.BytesSize,
			Percentage: pct,
		})
	}
	return vo.UserStorageStats{
		UsedBytes:  usedBytes,
		FreeBytes:  freeBytes,
		TotalBytes: totalBytes,
		Items:      items,
	}, nil
}

// ═══════════════════════════════════════════════════════════════
// 通用工具
// ═══════════════════════════════════════════════════════════════
