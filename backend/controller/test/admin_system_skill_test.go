package test

import (
	"fmt"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestAdminListSystemSkills(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminCreateSystemSkill(t *testing.T) {
	req := map[string]interface{}{
		"content": "---\nname: goraven-test-skill234\ndescription: 这是一个测试技能\n---\n# 测试技能\n\n你是一个测试助手。",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminGetSystemSkillDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminUpdateSystemSkill(t *testing.T) {
	req := map[string]interface{}{
		"content": "---\nname: goraven-test-skill\ndescription: 这是更新后的测试技能\n---\n# 更新后的测试技能\n\n你是一个更新后的测试助手。",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminUpdateSystemSkillStatus(t *testing.T) {
	req := map[string]interface{}{
		"status": 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills/2/status").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)

	req2 := map[string]interface{}{
		"status": 0,
	}
	str2, httpResp2 := requests.NewHTTPRequest(domain+"/api/admin/systemSkills/2/status").
		Put().
		SetJSONBody(req2).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str2, httpResp2)
}

func TestAdminDeleteSystemSkill(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills/2").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminCreateSystemSkillInvalidFormat(t *testing.T) {
	req := map[string]interface{}{
		"content": "这不是一个合法的技能格式",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminCreateSystemSkillInvalidName(t *testing.T) {
	req := map[string]interface{}{
		"content": "---\nname: invalid-skill\ndescription: 名称不合规\n---\n# 内容",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log(str, httpResp)
}

func TestAdminCreateSystemSkillDuplicate(t *testing.T) {
	content := "---\nname: goraven-dup-test\ndescription: 重复测试\n---\n# 内容"
	req := map[string]interface{}{
		"content": content,
	}

	str1, httpResp1 := requests.NewHTTPRequest(domain+"/api/admin/systemSkills").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log("first create:", str1, httpResp1)

	str2, httpResp2 := requests.NewHTTPRequest(domain+"/api/admin/systemSkills").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	t.Log("duplicate create:", str2, httpResp2)
}

func TestAdminSystemSkillNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/systemSkills/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).ToString()
	fmt.Println(str, httpResp)
}
