package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// TestGetMarketSkills 市场技能列表 GET /api/admin/marketSkills
func TestGetMarketSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMarketSkillsWithSearch 市场技能列表带搜索
func TestGetMarketSkillsWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills?search=sum").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMarketSkillsWithSourceFilter 市场技能列表按来源筛选
func TestGetMarketSkillsWithSourceFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills?source=custom_upload").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMarketSkillsWithStatusFilter 市场技能列表按状态筛选
func TestGetMarketSkillsWithStatusFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills?status=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMarketSkillDetail 市场技能详情 GET /api/admin/marketSkills/:id
func TestGetMarketSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdateMarketSkill 编辑市场技能 PUT /api/admin/marketSkills/:id
func TestUpdateMarketSkill(t *testing.T) {
	req := vo.AdminUpdateMarketSkillReq{
		Icon:        stringPtr("folder-git"),
		SortOrder:   intPtr(10),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdateMarketSkillStatus 上架/下架市场技能 PUT /api/admin/marketSkills/:id/status
func TestUpdateMarketSkillStatus(t *testing.T) {
	status := uint8(0)
	req := vo.AdminMarketSkillStatusReq{Status: &status}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5/status").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log("disable:", str, httpResp)

	enable := uint8(1)
	req2 := vo.AdminMarketSkillStatusReq{Status: &enable}
	str2, httpResp2 := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5/status").
		Put().
		SetJSONBody(req2).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log("enable:", str2, httpResp2)
}

// TestDeleteMarketSkill 删除市场技能 DELETE /api/admin/marketSkills/:id
func TestDeleteMarketSkill(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMarketSkillUsers 市场技能已安装用户列表 GET /api/admin/marketSkills/:id/users
func TestGetMarketSkillUsers(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/1/users").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestPublishMarketSkill 发布市场技能 POST /api/admin/marketSkills/publish
func TestPublishMarketSkill(t *testing.T) {
	req := vo.AdminPublishMarketSkillReq{
		UploadId:   "ba2805e64f394547b9777b9d69f5ca33",
		Icon:       "eino-guide",
		CategoryId: 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/publish").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestImportClawHubSkill 从 ClawHub 导入技能 POST /api/admin/marketSkills/import
func TestImportClawHubSkill(t *testing.T) {
	req := vo.AdminImportClawHubSkillReq{
		Slug:       "bloomberg-headlines",
		Icon:       "book-open-text",
		CategoryId: 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/import").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestSearchClawHub 搜索 ClawHub 技能 GET /api/admin/clawhub/search
func TestSearchClawHub(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/search?q=ppt&limit=10").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestExploreClawHub 浏览 ClawHub 技能列表 GET /api/admin/clawhub/explore
func TestExploreClawHub(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/explore?sort=trending").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestExploreClawHubByDownloads 按下载量浏览 ClawHub
func TestExploreClawHubByDownloads(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/explore?sort=downloads").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetClawHubSkillDetail ClawHub 技能详情 GET /api/admin/clawhub/skills/:slug
func TestGetClawHubSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/skills/summary").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMarketSkillNotFound 查询不存在的市场技能
func TestGetMarketSkillNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func stringPtr(s string) *string { return &s }
func intPtr(i int) *int          { return &i }

// ════════════════════════════════════════════════════════════════════════════
// 技能分类相关测试
// ════════════════════════════════════════════════════════════════════════════

// TestGetSkillCategories 技能分类列表 GET /api/admin/skillCategories
func TestGetSkillCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetAllSkillCategories 获取所有分类 GET /api/admin/skillCategories/all
func TestGetAllSkillCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories/all").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreateSkillCategory 创建技能分类 POST /api/admin/skillCategories
func TestCreateSkillCategory(t *testing.T) {
	req := vo.AdminCreateSkillCategoryReq{
		Name: "Development Tools",
		Icon: "code",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdateSkillCategory 编辑技能分类 PUT /api/admin/skillCategories/:id
func TestUpdateSkillCategory(t *testing.T) {
	req := vo.AdminUpdateSkillCategoryReq{
		Name: stringPtr("Updated Category"),
		Icon: stringPtr("wrench"),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestDeleteSkillCategory 删除技能分类 DELETE /api/admin/skillCategories/:id
func TestDeleteSkillCategory(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories/1").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
