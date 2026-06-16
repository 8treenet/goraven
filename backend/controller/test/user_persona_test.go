package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestGetUserPersonaCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/personaCategories").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetUserPersonaTemplates(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/personaTemplates").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetUserPersonaTemplatesWithCategoryFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/personaTemplates?categoryId=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetUserPersonaTemplateDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/personaTemplates/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetUserPersonaTemplateDetailNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/personaTemplates/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetUserPersonas(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateUserPersona(t *testing.T) {
	tmplId := 1
	req := vo.CreateUserPersonaReq{
		Name:		"代码审查助手",
		Icon:		"code-2",
		RoleInfo:	"你是一个专业的代码审查助手，擅长发现代码中的潜在问题并提供改进建议。请用简洁清晰的语言指出问题并给出修改方案。",
		CategoryId:	2,
		McpIds:		[]int{1, 3},
		SkillIds:	[]int{1, 2},
		AIModelId:	0,
		TemplateId:	&tmplId,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateUserPersona2(t *testing.T) {
	req := vo.CreateUserPersonaReq{
		Name:		"翻译助手",
		Icon:		"languages",
		RoleInfo:	"你是一个专业的翻译助手，精通中文、英文、日文等多种语言。请提供准确、流畅的翻译服务。",
		CategoryId:	3,
		McpIds:		[]int{},
		SkillIds:	[]int{3},
		AIModelId:	0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateUserPersona3(t *testing.T) {
	req := vo.CreateUserPersonaReq{
		Name:		"数据探索者",
		Icon:		"bar-chart-3",
		RoleInfo:	"你是一个专业的数据分析师，擅长从数据中发现洞察并提供可视化建议。",
		CategoryId:	5,
		McpIds:		[]int{4},
		SkillIds:	[]int{10, 11},
		AIModelId:	0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetUserPersonaDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdateUserPersona(t *testing.T) {
	name := "高级代码审查助手"
	icon := "shield-check"
	roleInfo := "你是一位资深的代码审查专家，精通多种编程语言和框架。请严格审查代码的安全性、性能和可维护性。"
	req := vo.UpdateUserPersonaReq{
		Name:		&name,
		Icon:		&icon,
		RoleInfo:	&roleInfo,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdateUserPersonaMcp(t *testing.T) {
	mcpIds := []int{1, 3, 6}
	req := vo.UpdateUserPersonaReq{
		McpIds: &mcpIds,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/3").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdateUserPersonaSkills(t *testing.T) {
	skillIds := []int{1, 2, 3}
	req := vo.UpdateUserPersonaReq{
		SkillIds: &skillIds,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/3").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetUserPersonaNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateUserPersonaMissingRoleInfo(t *testing.T) {
	req := vo.CreateUserPersonaReq{
		Name:		"空提示词角色",
		Icon:		"bot",
		RoleInfo:	"",
		CategoryId:	1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateUserPersonaMissingCategory(t *testing.T) {
	req := vo.CreateUserPersonaReq{
		Name:		"无分类角色",
		Icon:		"bot",
		RoleInfo:	"你是一个测试角色。",
		CategoryId:	0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeleteUserPersona(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/3").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeleteUserPersonaNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/personas/99999").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
