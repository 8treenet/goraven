package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestUserLogin(t *testing.T) {
	req := vo.UserLoginReq{
		Username: "admin",
		Password: "e10adc3949ba59abbe56e057f20f883e",
	}
	respBody, resp := requests.NewHTTPRequest(domain + "/api/user/login").Post().SetJSONBody(req).ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("login response:", respBody)
}

func TestGetUserInfo(t *testing.T) {
	respBody, resp := requests.NewHTTPRequest(domain+"/api/user").
		Get().SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("user info response:", respBody)
}

func TestUpdateProfile(t *testing.T) {
	token = "rvn_45014c60766f4df9970d1bec6362f84f"
	nickname := "测试昵称"
	req := vo.UserProfileReq{
		Nickname: &nickname,
	}
	respBody, resp := requests.NewHTTPRequest(domain+"/api/user/profile").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req).ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("update profile response:", respBody)

	email := "test@example.com"
	req2 := vo.UserProfileReq{
		Email: &email,
	}
	respBody2, resp2 := requests.NewHTTPRequest(domain+"/api/user/profile").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req2).ToString()
	if resp2.Error != nil {
		t.Log("error:", resp2.Error)
		return
	}
	t.Log("update email response:", respBody2)

	avatar := "/avatars/test.png"
	req3 := vo.UserProfileReq{
		Avatar: &avatar,
	}
	respBody3, resp3 := requests.NewHTTPRequest(domain+"/api/user/profile").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req3).ToString()
	if resp3.Error != nil {
		t.Log("error:", resp3.Error)
		return
	}
	t.Log("update avatar response:", respBody3)

	emptyEmail := "10000@qq.com"
	req4 := vo.UserProfileReq{
		Email: &emptyEmail,
	}
	respBody4, resp4 := requests.NewHTTPRequest(domain+"/api/user/profile").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req4).ToString()
	if resp4.Error != nil {
		t.Log("error:", resp4.Error)
		return
	}
	t.Log("set email response:", respBody4)
}

func TestChangePassword(t *testing.T) {
	req := vo.UserPasswordReq{
		CurrentPassword: "e10adc3949ba59abbe56e057f20f883e",
		NewPassword:     "dddd",
	}
	respBody, resp := requests.NewHTTPRequest(domain+"/api/user/password").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req).ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("change password (same) response:", respBody)
}

func TestUserLogout(t *testing.T) {
	token = "rvn_fb1aa9e24935423d9e0c4eed74c76e31"
	respBody, resp := requests.NewHTTPRequest(domain+"/api/user/logout").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("logout response:", respBody)
}
func TestUserDashboard(t *testing.T) {
	var rsp struct {
		Code int                 `json:"code"`
		Msg  string              `json:"msg"`
		Data vo.UserDashboardRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/dashboard?refresh=true").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("get user dashboard error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("get user dashboard failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}

	o := rsp.Data.Overview
	t.Logf("overview: todayTokens=%d weekTokens=%d totalSessions=%d newSessions=%d sparkline=%d",
		o.TodayTokens, o.WeekTokens, o.TotalSessions, o.NewSessions, len(o.Sparkline))

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

	t.Logf("storageStats: usedBytes=%d freeBytes=%d totalBytes=%d items=%d", rsp.Data.StorageStats.UsedBytes, rsp.Data.StorageStats.FreeBytes, rsp.Data.StorageStats.TotalBytes, len(rsp.Data.StorageStats.Items))
	for _, s := range rsp.Data.StorageStats.Items {
		t.Logf("  %s: bytes=%d pct=%.1f%%", s.Name, s.BytesSize, s.Percentage)
	}
}

func TestUserDashboardTokenTrend(t *testing.T) {
	var rsp struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data vo.TokenTrendRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/dashboard/tokenTrend?days=30&refresh=true").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("get user token trend error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("get user token trend failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}

	t.Logf("tokenTrend: %d points", len(rsp.Data.Items))
	if len(rsp.Data.Items) >= 2 {
		first := rsp.Data.Items[0]
		last := rsp.Data.Items[len(rsp.Data.Items)-1]
		t.Logf("  first: date=%s prompt=%d completion=%d", first.Date, first.PromptTokens, first.CompletionTokens)
		t.Logf("  last:  date=%s prompt=%d completion=%d", last.Date, last.PromptTokens, last.CompletionTokens)
		for i, p := range rsp.Data.Items {
			if p.PromptTokens > 0 || p.CompletionTokens > 0 {
				t.Logf("  #%d: date=%s prompt=%d completion=%d", i, p.Date, p.PromptTokens, p.CompletionTokens)
			}
		}
	}

	var rsp7 struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data vo.TokenTrendRsp `json:"data,omitempty"`
	}
	httpResp7 := requests.NewHTTPRequest(domain+"/api/dashboard/tokenTrend?days=7").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp7)
	if httpResp7.Error != nil {
		t.Fatal("get user token trend 7d error:", httpResp7.Error)
	}
	if rsp7.Code != 0 {
		t.Fatalf("get user token trend 7d failed: code=%d msg=%s", rsp7.Code, rsp7.Msg)
	}
	t.Logf("tokenTrend 7d: %d points", len(rsp7.Data.Items))

	var rspDef struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data vo.TokenTrendRsp `json:"data,omitempty"`
	}
	httpRespDef := requests.NewHTTPRequest(domain+"/api/dashboard/tokenTrend").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rspDef)
	if httpRespDef.Error != nil {
		t.Fatal("get user token trend default error:", httpRespDef.Error)
	}
	if rspDef.Code != 0 {
		t.Fatalf("get user token trend default failed: code=%d msg=%s", rspDef.Code, rspDef.Msg)
	}
	t.Logf("tokenTrend default: %d points", len(rspDef.Data.Items))
}
