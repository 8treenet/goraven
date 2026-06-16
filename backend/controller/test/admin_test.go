package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestAdminListUsers(t *testing.T) {
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
		Data	struct {
			List		[]vo.AdminUserItem	`json:"list"`
			TotalPage	int			`json:"totalPage"`
			TotalCount	int			`json:"totalCount"`
		}	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("list users error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("list users failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("list users success: total=%d users=%+v", rsp.Data.TotalCount, rsp.Data.List)
}

func TestAdminListUsersWithSearch(t *testing.T) {
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
		Data	struct {
			List		[]vo.AdminUserItem	`json:"list"`
			TotalPage	int			`json:"totalPage"`
			TotalCount	int			`json:"totalCount"`
		}	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users?search=admin").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("list users with search error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("list users with search failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("search users success: total=%d users=%+v", rsp.Data.TotalCount, rsp.Data.List)
}

func TestAdminListUsersWithRoleFilter(t *testing.T) {
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
		Data	struct {
			List		[]vo.AdminUserItem	`json:"list"`
			TotalPage	int			`json:"totalPage"`
			TotalCount	int			`json:"totalCount"`
		}	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users?role=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("list users with role filter error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("list users with role filter failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("filter admin users success: total=%d users=%+v", rsp.Data.TotalCount, rsp.Data.List)
}

func TestAdminGetUserDetail(t *testing.T) {
	userId := "999"
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.AdminUserDetailRsp	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users/"+userId).
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("get user detail error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("get user detail failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("get user detail success: userId=%s username=%s nickname=%s role=%d status=%d",
		rsp.Data.UserId, rsp.Data.Username, rsp.Data.Nickname, rsp.Data.Role, rsp.Data.Status)
}

func TestAdminBatchGetUsers(t *testing.T) {
	req := vo.AdminBatchUserReq{
		UserIds: []string{"999", "8dde3252e54d43ecbc0d981b038c311c"},
	}
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
		Data	struct {
			List []vo.AdminUserItem `json:"list"`
		}	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users/batch").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("batch get users error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("batch get users failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	if len(rsp.Data.List) == 0 {
		t.Fatal("batch get users: no users returned")
	}
	t.Logf("batch get users success: count=%d users=%+v", len(rsp.Data.List), rsp.Data.List)
}

func TestAdminCreateUser(t *testing.T) {
	req := vo.AdminCreateUserReq{
		Username:	"testuser_333" + t.Name(),
		Password:	"e10adc3949ba59abbe56e057f20f883e",
		Nickname:	"Test User",
		Role:		1,
	}
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("create user error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("create user failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("create user success: username=%s", req.Username)
}

func TestAdminUpdateUser(t *testing.T) {
	req := vo.AdminUpdateUserReq{
		Nickname: "啊哈哈哈",
	}
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users/b68af727c9eb49a4b262e3b430db623f").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("update user error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("update user failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("update user success: nickname=%s", req.Nickname)
}

func TestAdminUpdateUserStatus(t *testing.T) {
	status := uint8(0)
	req := vo.AdminUpdateUserReq{
		Nickname:	"Test User",
		Status:		&status,
	}
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users/b68af727c9eb49a4b262e3b430db623f").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("update user status error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("update user status failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("update user status success")

	var rsp2 struct {
		Code int `json:"code"`
	}
	enableStatus := uint8(1)
	req2 := vo.AdminUpdateUserReq{
		Status: &enableStatus,
	}
	httpResp2 := requests.NewHTTPRequest(domain+"/api/admin/users/b68af727c9eb49a4b262e3b430db623f").
		Put().
		SetJSONBody(req2).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp2)
	if httpResp2.Error != nil {
		t.Fatal("restore user status error:", httpResp2.Error)
	}
	if rsp2.Code != 0 {
		t.Fatalf("restore user status failed: code=%d", rsp2.Code)
	}
	t.Logf("restore user status success")
}

func TestAdminResetPassword(t *testing.T) {
	req := vo.AdminResetPasswordReq{
		Password: "e10adc3949ba59abbe56e057f20f883e",
	}
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/users/b68af727c9eb49a4b262e3b430db623f/reset-password").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("reset password error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("reset password failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Log("reset password success")
}

func TestAdminDeleteSuperAdminFails(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/users/5a1f5bd926c44f5ebd3a51d8865f475f").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminSystemInfo(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemInfo").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log("systemInfo:", str, httpResp)
}

func TestAdminSystemInfoForceRefresh(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemInfo?forceRefresh=true").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log("systemInfo forceRefresh:", str, httpResp)
}

func TestAdminGetSettings(t *testing.T) {
	var rsp struct {
		Code	int	`json:"code"`
		Msg	string	`json:"msg"`
		Data	struct {
			Groups []vo.AdminSettingGroup `json:"groups"`
		}	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/admin/settings").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("get settings error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("get settings failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	if len(rsp.Data.Groups) == 0 {
		t.Fatal("get settings: no groups returned")
	}
	for _, g := range rsp.Data.Groups {
		t.Logf("group: name=%s displayName=%s settings=%d", g.Name, g.DisplayName, len(g.Settings))
		for _, s := range g.Settings {
			t.Logf("  setting: key=%s value=%s valueType=%s inputType=%s displayName=%s", s.Key, s.Value, s.ValueType, s.InputType, s.DisplayName)
		}
	}
}

func TestAdminUpdateSettings(t *testing.T) {

	req := vo.AdminUpdateSettingsReq{
		Settings: []vo.AdminSettingUpdateItem{
			{Key: "general.domain", Value: "https://goraven.dev"},
		},
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/settings").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	if httpResp.Error != nil {
		t.Fatal("update settings error:", httpResp.Error)
	}
	t.Log("update settings:", str, httpResp)
}
