package service

import (
	"fmt"

	"raven/backend/infra"
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/core/sandbox"
	"raven/util"

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
	Worker		freedom.Worker
	DashboardRepo	*repository.DashboardRepository
}

func (service *DashboardService) GetAdminDashboard(refresh bool) (*vo.AdminDashboardRsp, error) {
	const cacheKey = "dashboard:admin"
	if refresh {
		service.DashboardRepo.InvalidateDashboardCache(cacheKey)
	}

	rsp := &vo.AdminDashboardRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	last7 := util.GenLastNDates(7)
	thisWeek := util.GenThisWeekDates()
	last30 := util.GenLastNDates(30)

	overview, err := service.buildAdminOverview(last7, thisWeek)
	if err != nil {
		return nil, err
	}
	rsp.Overview = *overview

	tokenTrend, err := service.DashboardRepo.GetTokenTrend(last30)
	if err != nil {
		return nil, err
	}
	rsp.TokenTrend = tokenTrend

	modelUsage, err := service.DashboardRepo.GetModelUsage()
	if err != nil {
		return nil, err
	}
	rsp.ModelUsage = modelUsage

	userRank, err := service.DashboardRepo.GetUserTokenRank()
	if err != nil {
		return nil, err
	}
	rsp.UserTokenRank = userRank

	activeTrend, err := service.DashboardRepo.GetActiveUserTrend(last30)
	if err != nil {
		return nil, err
	}
	rsp.ActiveTrend = activeTrend

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

	rsp.Alerts = buildAlerts(overview.TodayTokens, tokenTrend)

	service.DashboardRepo.SetDashboardCache(cacheKey, rsp)
	return rsp, nil
}

func (service *DashboardService) GetAdminTokenTrend(days int, refresh bool) (*vo.TokenTrendRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:admin:tokenTrend:%d", days)

	if refresh {
		service.DashboardRepo.InvalidateDashboardCache(cacheKey)
	}
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

func (service *DashboardService) GetAdminActiveUserTrend(days int, refresh bool) (*vo.ActiveUserTrendRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:admin:activeUsers:%d", days)

	if refresh {
		service.DashboardRepo.InvalidateDashboardCache(cacheKey)
	}

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

func (service *DashboardService) GetUserDashboard(userId string, refresh bool) (*vo.UserDashboardRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:user:%s", userId)

	if refresh {
		service.DashboardRepo.InvalidateDashboardCache(cacheKey)
	}

	rsp := &vo.UserDashboardRsp{}
	if service.DashboardRepo.GetDashboardCache(cacheKey, rsp) {
		return rsp, nil
	}

	last7 := util.GenLastNDates(7)
	thisWeek := util.GenThisWeekDates()
	last30 := util.GenLastNDates(30)

	overview, err := service.buildUserOverview(userId, last7, thisWeek)
	if err != nil {
		return nil, err
	}
	rsp.Overview = *overview

	tokenTrend, err := service.DashboardRepo.GetUserTokenTrend(userId, last30)
	if err != nil {
		return nil, err
	}
	rsp.TokenTrend = tokenTrend

	modelUsage, err := service.DashboardRepo.GetUserModelUsage(userId)
	if err != nil {
		return nil, err
	}
	rsp.ModelUsage = modelUsage

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
	return rsp, nil
}

func (service *DashboardService) GetUserTokenTrend(userId string, days int, refresh bool) (*vo.TokenTrendRsp, error) {
	cacheKey := fmt.Sprintf("dashboard:user:%s:tokenTrend:%d", userId, days)

	if refresh {
		service.DashboardRepo.InvalidateDashboardCache(cacheKey)
	}

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
			Name:		u.Name,
			BytesSize:	u.BytesSize,
			Percentage:	pct,
		})
	}
	return vo.UserStorageStats{
		UsedBytes:	usedBytes,
		FreeBytes:	freeBytes,
		TotalBytes:	totalBytes,
		Items:		items,
	}, nil
}

func buildAlerts(todayTokens int64, tokenTrend []vo.TokenTrendItem) []vo.DashboardAlert {
	var alerts []vo.DashboardAlert

	if len(tokenTrend) > 0 && todayTokens > 0 {
		var totalTokens int64
		for _, t := range tokenTrend {
			totalTokens += t.PromptTokens + t.CompletionTokens
		}
		avgDaily := totalTokens / int64(len(tokenTrend))
		if avgDaily > 0 && todayTokens > avgDaily*2 {
			alerts = append(alerts, vo.DashboardAlert{
				Level:		"warning",
				Message:	"当日 Token 消耗异常偏高，已超过近期日均值的 2 倍",
			})
		}
	}

	return alerts
}
