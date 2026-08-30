package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// ════════════════════════════════════════════════════════════════════════════
// 市场技能（用户端）
// ════════════════════════════════════════════════════════════════════════════

// TestUserListMarketSkills 用户市场技能列表 GET /api/skills/market
func TestUserListMarketSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserListMarketSkillsWithSearch 用户市场技能列表带搜索
func TestUserListMarketSkillsWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?search=Doc").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserListMarketSkillsWithCategoryFilter 用户市场技能列表按分类筛选
func TestUserListMarketSkillsWithCategoryFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?categoryId=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserListMarketSkillsWithSourceFilter 用户市场技能列表按来源筛选
func TestUserListMarketSkillsWithSourceFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?source=clawhub").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserListMarketSkillsPaginated 用户市场技能列表分页
func TestUserListMarketSkillsPaginated(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?page=1&pageSize=5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserGetMarketSkillDetail 用户市场技能详情 GET /api/skills/market/:id
func TestUserGetMarketSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserGetMarketSkillDetailNotFound 查询不存在的市场技能
func TestUserGetMarketSkillDetailNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// ════════════════════════════════════════════════════════════════════════════
// 技能分类（用户端）
// ════════════════════════════════════════════════════════════════════════════

// TestUserListSkillCategories 用户技能分类列表 GET /api/skills/categories
func TestUserListSkillCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/categories").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// ════════════════════════════════════════════════════════════════════════════
// 用户技能 CRUD
// ════════════════════════════════════════════════════════════════════════════

// TestUserListUserSkills 用户已安装技能列表 GET /api/skills/user
func TestUserListUserSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserListUserSkillsWithSearch 用户技能列表带搜索
func TestUserListUserSkillsWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user?search=docker").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserListUserSkillsWithStatusFilter 用户技能列表按安装状态筛选
func TestUserListUserSkillsWithStatusFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user?status=2").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserListUserSkillsPaginated 用户技能列表分页
func TestUserListUserSkillsPaginated(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user?page=1&pageSize=5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserGetUserSkillDetail 用户技能详情 GET /api/skills/user/:id
func TestUserGetUserSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/7").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserGetUserSkillDetailNotFound 查询不存在的用户技能
func TestUserGetUserSkillDetailNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// ════════════════════════════════════════════════════════════════════════════
// 安装技能
// ════════════════════════════════════════════════════════════════════════════

// TestUserInstallSkill 安装市场技能 POST /api/skills/install
func TestUserInstallSkill(t *testing.T) {
	req := vo.UserSkillInstallReq{
		SkillId: 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/install").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserInstallSkillAlreadyInstalled 安装已安装的技能（应报错）
func TestUserInstallSkillAlreadyInstalled(t *testing.T) {
	req := vo.UserSkillInstallReq{
		SkillId: 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/install").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserInstallSkillNotFound 安装不存在的市场技能
func TestUserInstallSkillNotFound(t *testing.T) {
	req := vo.UserSkillInstallReq{
		SkillId: 99999,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/install").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// ════════════════════════════════════════════════════════════════════════════
// 编辑用户技能
// ════════════════════════════════════════════════════════════════════════════

// TestUserUpdateUserSkill 编辑用户技能（图标+分类） PUT /api/skills/user/:id
func TestUserUpdateUserSkill(t *testing.T) {
	icon := "globe"
	categoryId := 2
	req := vo.UserSkillUpdateReq{
		Icon:       &icon,
		CategoryId: &categoryId,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/9").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserUpdateUserSkillOnlyIcon 仅编辑图标
func TestUserUpdateUserSkillOnlyIcon(t *testing.T) {
	icon := "folder-git"
	req := vo.UserSkillUpdateReq{
		Icon: &icon,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/9").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserUpdateUserSkillOnlyCategory 仅编辑分类
func TestUserUpdateUserSkillOnlyCategory(t *testing.T) {
	categoryId := 1
	req := vo.UserSkillUpdateReq{
		CategoryId: &categoryId,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/9").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserUpdateUserSkillNotFound 编辑不存在的用户技能
func TestUserUpdateUserSkillNotFound(t *testing.T) {
	icon := "globe"
	req := vo.UserSkillUpdateReq{
		Icon: &icon,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/99999").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// ════════════════════════════════════════════════════════════════════════════
// 刷新 & 重试 & 删除
// ════════════════════════════════════════════════════════════════════════════

// TestUserRefreshUserSkills 刷新同步用户技能 POST /api/skills/user/refresh
func TestUserRefreshUserSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/refresh").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserRetryInstallSkill 重试安装失败技能 PUT /api/skills/user/:id/retry
func TestUserRetryInstallSkill(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/1/retry").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserRetryInstallSkillNotFound 重试安装不存在的用户技能
func TestUserRetryInstallSkillNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/99999/retry").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserDeleteUserSkill 删除用户技能 DELETE /api/skills/user/:id
func TestUserDeleteUserSkill(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/8").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUserDeleteUserSkillNotFound 删除不存在的用户技能
func TestUserDeleteUserSkillNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/99999").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// ════════════════════════════════════════════════════════════════════════════
// 精简技能列表
// ════════════════════════════════════════════════════════════════════════════

// TestUserGetSimpleSkills 获取用户可选技能精简列表 GET /api/skills/simpleSkills
func TestUserGetSimpleSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/simpleSkills").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
