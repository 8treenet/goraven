package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestGetMarketSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetMarketSkillsWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills?search=sum").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetMarketSkillsWithSourceFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills?source=custom_upload").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetMarketSkillsWithStatusFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills?status=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetMarketSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdateMarketSkill(t *testing.T) {
	req := vo.AdminUpdateMarketSkillReq{
		Icon:		stringPtr("folder-git"),
		SortOrder:	intPtr(10),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

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

func TestDeleteMarketSkill(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/5").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetMarketSkillUsers(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/1/users").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestPublishMarketSkill(t *testing.T) {
	req := vo.AdminPublishMarketSkillReq{
		UploadId:	"ba2805e64f394547b9777b9d69f5ca33",
		Icon:		"eino-guide",
		CategoryId:	1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/publish").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestImportClawHubSkill(t *testing.T) {
	req := vo.AdminImportClawHubSkillReq{
		Slug:		"bloomberg-headlines",
		Icon:		"book-open-text",
		CategoryId:	1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/import").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestSearchClawHub(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/search?q=ppt&limit=10").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestExploreClawHub(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/explore?sort=trending").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestExploreClawHubByDownloads(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/explore?sort=downloads").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetClawHubSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/clawhub/skills/summary").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetMarketSkillNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/marketSkills/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func stringPtr(s string) *string	{ return &s }
func intPtr(i int) *int			{ return &i }

func TestGetSkillCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetAllSkillCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories/all").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateSkillCategory(t *testing.T) {
	req := vo.AdminCreateSkillCategoryReq{
		Name:	"Development Tools",
		Icon:	"code",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdateSkillCategory(t *testing.T) {
	req := vo.AdminUpdateSkillCategoryReq{
		Name:	stringPtr("Updated Category"),
		Icon:	stringPtr("wrench"),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeleteSkillCategory(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/skillCategories/1").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
