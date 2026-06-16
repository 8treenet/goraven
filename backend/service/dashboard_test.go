package service_test

import (
	"raven/backend/repository"
	"raven/config"
	"raven/util"
	unit_test "raven/util/unit"
	"testing"
	"time"

	pkg_dashboard "raven/backend/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDashboard_CheckDailyStatsData(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.DashboardRepository
	unitTest.FetchRepository(&repo)

	dateRange := util.GenLastNDates(90)
	items, err := repo.GetTokenTrend(dateRange)
	if err != nil {
		t.Fatalf("GetTokenTrend error: %v", err)
	}
	t.Logf("GetTokenTrend(90d) returned %d points", len(items))

	totalPrompt, totalCompletion, zeroDays := int64(0), int64(0), 0
	for _, item := range items {
		totalPrompt += item.PromptTokens
		totalCompletion += item.CompletionTokens
		if item.PromptTokens == 0 && item.CompletionTokens == 0 {
			zeroDays++
		} else {
			t.Logf("  date=%s prompt=%d completion=%d", item.Date, item.PromptTokens, item.CompletionTokens)
		}
	}
	t.Logf("zeroDays=%d/%d (%.0f%%)", zeroDays, len(items), float64(zeroDays)/float64(len(items))*100)
	t.Logf("totalPrompt=%d totalCompletion=%d", totalPrompt, totalCompletion)
}

func TestDashboard_GetAdminDashboard(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var svc *pkg_dashboard.DashboardService
	unitTest.FetchService(&svc)

	rsp, err := svc.GetAdminDashboard(true)
	if err != nil {
		t.Fatalf("GetAdminDashboard error: %v", err)
	}
	t.Logf("overview: activeUsers=%d sessions=%d newSessions=%d weekTokens=%d todayTokens=%d models=%d",
		rsp.Overview.ActiveUsers, rsp.Overview.TotalSessions, rsp.Overview.NewSessions,
		rsp.Overview.WeekTokens, rsp.Overview.TodayTokens, rsp.Overview.EnabledModels)
	t.Logf("tokenTrend: %d points", len(rsp.TokenTrend))

	hasData := false
	for _, item := range rsp.TokenTrend {
		if item.PromptTokens > 0 || item.CompletionTokens > 0 {
			t.Logf("  NONZERO date=%s prompt=%d completion=%d", item.Date, item.PromptTokens, item.CompletionTokens)
			hasData = true
		}
	}
	if !hasData {
		t.Log("  (all zero)")
	}
	t.Logf("modelUsage: %d items", len(rsp.ModelUsage))
	for _, m := range rsp.ModelUsage {
		t.Logf("  %s: tokens=%d pct=%.1f%%", m.ModelName, m.TokenCount, m.Percentage)
	}
}

func TestDashboard_GetAdminTokenTrend(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var svc *pkg_dashboard.DashboardService
	unitTest.FetchService(&svc)

	for _, days := range []int{7, 30, 90} {
		rsp, err := svc.GetAdminTokenTrend(days, true)
		if err != nil {
			t.Fatalf("GetAdminTokenTrend(%d) error: %v", days, err)
		}
		totalPrompt, totalCompletion := int64(0), int64(0)
		for _, item := range rsp.Items {
			totalPrompt += item.PromptTokens
			totalCompletion += item.CompletionTokens
		}
		t.Logf("GetAdminTokenTrend(%dd): %d points, totalPrompt=%d totalCompletion=%d",
			days, len(rsp.Items), totalPrompt, totalCompletion)
	}
}

func TestDashboard_CheckAddDailyStats(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.DailyStatsRepository
	unitTest.FetchRepository(&repo)

	err := repo.AddDailyStats("test_user_99", 100, 200, 10)
	if err != nil {
		t.Fatalf("AddDailyStats error: %v", err)
	}
	t.Log("AddDailyStats(test_user_99, 100, 200, 10) OK")
}

func TestDashboard_RawTableQuery(t *testing.T) {
	cfg := config.Get()
	dbPath := cfg.Paths.SqliteDB
	t.Logf("DB path: %s", dbPath)

	rawDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	var totalCount int64
	rawDB.Table("user_daily_stats").Count(&totalCount)
	t.Logf("user_daily_stats total rows: %d", totalCount)

	type rawRow struct {
		UserId			string
		StatDate		string
		PromptTokens		int64
		CompletionTokens	int64
		RoundCount		int64
	}
	var todayRows []rawRow
	rawDB.Raw(
		"SELECT userId, statDate, promptTokens, completionTokens, roundCount FROM user_daily_stats WHERE statDate >= ? ORDER BY statDate DESC LIMIT 10",
		time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
	).Scan(&todayRows)
	for _, r := range todayRows {
		t.Logf("  userId=%s statDate=%s prompt=%d completion=%d round=%d",
			r.UserId, r.StatDate, r.PromptTokens, r.CompletionTokens, r.RoundCount)
	}

	type aggRow struct {
		StatDate		string
		PromptTokens		int64
		CompletionTokens	int64
	}
	var aggRows []aggRow
	rawDB.Raw(
		"SELECT statDate, COALESCE(SUM(promptTokens), 0) AS promptTokens, COALESCE(SUM(completionTokens), 0) AS completionTokens FROM user_daily_stats GROUP BY statDate ORDER BY statDate DESC LIMIT 10",
	).Scan(&aggRows)
	t.Logf("aggregated rows: %d", len(aggRows))
	for _, r := range aggRows {
		t.Logf("  aggregated: statDate=%s prompt=%d completion=%d",
			r.StatDate, r.PromptTokens, r.CompletionTokens)
	}
}

func TestDashboard_DataFlow(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var dsRepo *repository.DailyStatsRepository
	unitTest.FetchRepository(&dsRepo)

	var dashRepo *repository.DashboardRepository
	unitTest.FetchRepository(&dashRepo)

	today := util.StartOfToday().Format("2006-01-02")
	t.Logf("today=%s", today)

	if err := dsRepo.AddDailyStats("test_user_99", 300, 400, 20); err != nil {
		t.Fatalf("AddDailyStats error: %v", err)
	}
	t.Logf("AddDailyStats done (300+400)")

	items, err := dashRepo.GetTokenTrend([]string{today})
	if err != nil {
		t.Fatalf("GetTokenTrend error: %v", err)
	}
	t.Logf("GetTokenTrend(%s): %d items", today, len(items))
	for _, item := range items {
		t.Logf("  date=%s prompt=%d completion=%d", item.Date, item.PromptTokens, item.CompletionTokens)
		if item.PromptTokens > 0 || item.CompletionTokens > 0 {
			t.Logf("  ★ DATA FOUND")
		}
	}
}
