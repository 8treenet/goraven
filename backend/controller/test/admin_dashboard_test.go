package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// TestAdminDashboard 管理员仪表盘聚合数据 GET /api/admin/dashboard
func TestAdminDashboard(t *testing.T) {

	var rsp struct {
		Code int                  `json:"code"`
		Msg  string               `json:"msg"`
		Data vo.AdminDashboardRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/dashboard").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("get dashboard error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("get dashboard failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}

	o := rsp.Data.Overview
	t.Logf("overview: activeUsers=%d diff=%.1f%% sessions=%d newSessions=%d weekTokens=%d todayTokens=%d models=%d sparkline=%d",
		o.ActiveUsers, o.ActiveUsersDiff, o.TotalSessions, o.NewSessions, o.WeekTokens, o.TodayTokens, o.EnabledModels, len(o.Sparkline))

	t.Logf("skillUsageRank: %d items", len(rsp.Data.SkillUsageRank))
	for _, s := range rsp.Data.SkillUsageRank {
		t.Logf("  %s: count=%d", s.Name, s.Count)
	}

	t.Logf("mcpUsageRank: %d items", len(rsp.Data.McpUsageRank))
	for _, m := range rsp.Data.McpUsageRank {
		t.Logf("  %s: count=%d", m.Name, m.Count)
	}

	t.Logf("toolUsageRank: %d items", len(rsp.Data.ToolUsageRank))
	for _, item := range rsp.Data.ToolUsageRank {
		t.Logf("  %s: count=%d", item.Name, item.Count)
	}

}

// TestAdminDashboardTokenTrend Token 趋势 GET /api/admin/dashboard/tokenTrend?days=30
func TestAdminDashboardTokenTrend(t *testing.T) {
	var rsp struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data vo.TokenTrendRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/dashboard/tokenTrend?days=30").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("get token trend error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("get token trend failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}

	t.Logf("tokenTrend: %d points", len(rsp.Data.Items))
	if len(rsp.Data.Items) >= 2 {
		first := rsp.Data.Items[0]
		last := rsp.Data.Items[len(rsp.Data.Items)-1]
		t.Logf("  first: date=%s prompt=%d completion=%d", first.Date, first.PromptTokens, first.CompletionTokens)
		t.Logf("  last:  date=%s prompt=%d completion=%d", last.Date, last.PromptTokens, last.CompletionTokens)
	}

	// 测试 7 天粒度
	var rsp7 struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data vo.TokenTrendRsp `json:"data,omitempty"`
	}
	httpResp7 := requests.NewHTTPRequest(domain+"/api/admin/dashboard/tokenTrend?days=7").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp7)
	if httpResp7.Error != nil {
		t.Fatal("get token trend 7d error:", httpResp7.Error)
	}
	if rsp7.Code != 0 {
		t.Fatalf("get token trend 7d failed: code=%d msg=%s", rsp7.Code, rsp7.Msg)
	}
	t.Logf("tokenTrend 7d: %d points", len(rsp7.Data.Items))

	// 测试缺省参数
	var rspDef struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data vo.TokenTrendRsp `json:"data,omitempty"`
	}
	httpRespDef := requests.NewHTTPRequest(domain+"/api/admin/dashboard/tokenTrend").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rspDef)
	if httpRespDef.Error != nil {
		t.Fatal("get token trend default error:", httpRespDef.Error)
	}
	if rspDef.Code != 0 {
		t.Fatalf("get token trend default failed: code=%d msg=%s", rspDef.Code, rspDef.Msg)
	}
	t.Logf("tokenTrend default: %d points", len(rspDef.Data.Items))
}

// TestAdminDashboardActiveUsers 活跃用户趋势 GET /api/admin/dashboard/activeUsers?days=30
func TestAdminDashboardActiveUsers(t *testing.T) {
	var rsp struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data vo.ActiveUserTrendRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/dashboard/activeUsers?days=30").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("get active users error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("get active users failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}

	t.Logf("activeUsers: %d points", len(rsp.Data.Items))
	if len(rsp.Data.Items) >= 2 {
		first := rsp.Data.Items[0]
		last := rsp.Data.Items[len(rsp.Data.Items)-1]
		t.Logf("  first: date=%s count=%d", first.Date, first.Count)
		t.Logf("  last:  date=%s count=%d", last.Date, last.Count)
	}

	// 测试 7 天粒度
	var rsp7 struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data vo.ActiveUserTrendRsp `json:"data,omitempty"`
	}
	httpResp7 := requests.NewHTTPRequest(domain+"/api/admin/dashboard/activeUsers?days=7").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp7)
	if httpResp7.Error != nil {
		t.Fatal("get active users 7d error:", httpResp7.Error)
	}
	if rsp7.Code != 0 {
		t.Fatalf("get active users 7d failed: code=%d msg=%s", rsp7.Code, rsp7.Msg)
	}
	t.Logf("activeUsers 7d: %d points", len(rsp7.Data.Items))

	// 测试缺省参数
	var rspDef struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data vo.ActiveUserTrendRsp `json:"data,omitempty"`
	}
	httpRespDef := requests.NewHTTPRequest(domain+"/api/admin/dashboard/activeUsers").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rspDef)
	if httpRespDef.Error != nil {
		t.Fatal("get active users default error:", httpRespDef.Error)
	}
	if rspDef.Code != 0 {
		t.Fatalf("get active users default failed: code=%d msg=%s", rspDef.Code, rspDef.Msg)
	}
	t.Logf("activeUsers default: %d points", len(rspDef.Data.Items))
}
