package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestUserListMarketSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListMarketSkillsWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?search=Doc").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListMarketSkillsWithCategoryFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?categoryId=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListMarketSkillsWithSourceFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?source=clawhub").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListMarketSkillsPaginated(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market?page=1&pageSize=5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserGetMarketSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserGetMarketSkillDetailNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/market/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListSkillCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/categories").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListUserSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListUserSkillsWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user?search=docker").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListUserSkillsWithStatusFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user?status=2").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserListUserSkillsPaginated(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user?page=1&pageSize=5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserGetUserSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/7").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserGetUserSkillDetailNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

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

func TestUserRefreshUserSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/refresh").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserRetryInstallSkill(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/1/retry").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserRetryInstallSkillNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/99999/retry").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserDeleteUserSkill(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/8").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserDeleteUserSkillNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/user/99999").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUserGetSimpleSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/skills/simpleSkills").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
