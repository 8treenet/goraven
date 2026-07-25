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

type DashboardService struct {
	Worker        freedom.Worker
	DashboardRepo *repository.DashboardRepository
}

func (service *DashboardService) GetAdminDashboard() (*vo.AdminDashboardRsp, error) {
	const cacheKey = "dashboard:admin"

	rsp := &vo.AdminDashboardRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	last7 := util.GenLastNDates(7)
	thisWeek := util.GenThisWeekDates()

	overview, err := service.buildAdminOverview(last7, thisWeek)
	if err != nil {
		return nil, err
	}
	rsp.Overview = *overview

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

func (service *DashboardService) buildAdminOverview(last7, thisWeek []string) (*vo.AdminDashboardOverview, error) {
	overview := &vo.AdminDashboardOverview{}

	activeUsers, err := service.DashboardRepo.GetActiveUsersInRange(7, 0)
	if err != nil {
		return nil, err
	}
	overview.ActiveUsers = int(activeUsers)

	prevActive, err := service.DashboardRepo.GetActiveUsersInRange(14, 7)
	if err != nil {
		return nil, err
	}
	if prevActive > 0 {
		overview.ActiveUsersDiff = float64(activeUsers-prevActive) / float64(prevActive) * 100
	}

	totalSessions, newThisWeek, err := service.DashboardRepo.GetSessionCounts()
	if err != nil {
		return nil, err
	}
	overview.TotalSessions = totalSessions
	overview.NewSessions = newThisWeek

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

	modelCount, err := service.DashboardRepo.GetEnabledModelCount()
	if err != nil {
		return nil, err
	}
	overview.EnabledModels = modelCount

	sparkline, err := service.DashboardRepo.GetSparkline(last7)
	if err != nil {
		return nil, err
	}
	overview.Sparkline = sparkline

	return overview, nil
}

func (service *DashboardService) GetUserDashboard(userId string) (*vo.UserDashboardRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:user:%s", userId)

	rsp := &vo.UserDashboardRsp{}
	cacheHit := service.DashboardRepo.GetDashboardCache(cacheKey, rsp)

	if !cacheHit {
		last7 := util.GenLastNDates(7)
		thisWeek := util.GenThisWeekDates()

		overview, err := service.buildUserOverview(userId, last7, thisWeek)
		if err != nil {
			return nil, err
		}
		rsp.Overview = *overview

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

		storageStats, err := service.buildStorageStats(userId)
		if err != nil {
			return nil, err
		}
		rsp.StorageStats = storageStats

		service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	}

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

func (service *DashboardService) buildUserOverview(userId string, last7, thisWeek []string) (*vo.UserDashboardOverview, error) {
	overview := &vo.UserDashboardOverview{}

	totalSessions, newThisWeek, err := service.DashboardRepo.GetUserSessionCounts(userId)
	if err != nil {
		return nil, err
	}
	overview.TotalSessions = totalSessions
	overview.NewSessions = newThisWeek

	prompt, completion, err := service.DashboardRepo.GetUserTokensInDateRange(userId, thisWeek)
	if err != nil {
		return nil, err
	}
	overview.WeekTokens = prompt + completion

	todayTokens, err := service.DashboardRepo.GetUserTodayTokens(userId)
	if err != nil {
		return nil, err
	}
	overview.TodayTokens = todayTokens

	totalTokens, err := service.DashboardRepo.GetUserTotalTokens(userId)
	if err != nil {
		return nil, err
	}
	overview.TotalTokens = totalTokens

	sparkline, err := service.DashboardRepo.GetUserSparkline(userId, last7)
	if err != nil {
		return nil, err
	}
	overview.Sparkline = sparkline

	return overview, nil
}

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
