package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestGetPersonaTemplates(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPersonaTemplatesWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates?search=编程").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPersonaTemplatesWithCategoryFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates?categoryId=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreatePersonaTemplate(t *testing.T) {
	req := vo.AdminCreatePersonaTemplateReq{
		Name:		"编程助手",
		Icon:		"code",
		Description:	"擅长代码编写、调试和架构设计",
		RoleInfo:	"你是一个专业的编程助手，擅长多种编程语言和技术栈。帮助用户编写高质量代码、调试问题并提供架构建议。",
		CategoryId:	7,
		SortOrder:	0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreatePersonaTemplate2(t *testing.T) {
	req := vo.AdminCreatePersonaTemplateReq{
		Name:		"翻译专家",
		Icon:		"languages",
		Description:	"精通多语言翻译",
		RoleInfo:	"你是一个专业的翻译专家，精通中英日韩等多种语言。帮助用户进行准确的翻译，保持原文的语气和风格。",
		CategoryId:	1,
		SortOrder:	0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPersonaTemplateDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdatePersonaTemplate(t *testing.T) {
	req := vo.AdminUpdatePersonaTemplateReq{
		Name:		stringPtr("高级编程助手"),
		Icon:		stringPtr("code-2"),
		Description:	stringPtr("擅长全栈开发和系统架构"),
		SortOrder:	intPtr(1),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdatePersonaTemplateRoleInfo(t *testing.T) {
	req := vo.AdminUpdatePersonaTemplateReq{
		RoleInfo: stringPtr("你是一个资深全栈工程师，精通前后端开发、数据库设计和系统架构。用清晰简洁的方式解答技术问题。"),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPersonaTemplateNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreatePersonaTemplateMissingRoleInfo(t *testing.T) {
	req := vo.AdminCreatePersonaTemplateReq{
		Name:		"空提示词模板",
		Icon:		"bot",
		RoleInfo:	"",
		CategoryId:	1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeletePersonaTemplate(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/2").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPersonaCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetAllPersonaCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/all").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPersonaCategoryDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreatePersonaCategory(t *testing.T) {
	req := vo.AdminCreatePersonaCategoryReq{
		Name:	"自定义分类",
		Icon:	"sparkles",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdatePersonaCategory(t *testing.T) {
	req := vo.AdminUpdatePersonaCategoryReq{
		Name:	stringPtr("更新分类名"),
		Icon:	stringPtr("star"),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/7").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeletePersonaCategory(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/8").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPersonaCategoryNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
