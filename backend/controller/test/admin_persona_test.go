package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// ════════════════════════════════════════════════════════════════════════════
// 角色模板相关测试
// ════════════════════════════════════════════════════════════════════════════

// TestGetPersonaTemplates 角色模板列表 GET /api/admin/personaTemplates
func TestGetPersonaTemplates(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetPersonaTemplatesWithSearch 角色模板列表带搜索
func TestGetPersonaTemplatesWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates?search=编程").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetPersonaTemplatesWithCategoryFilter 角色模板列表按分类筛选
func TestGetPersonaTemplatesWithCategoryFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates?categoryId=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreatePersonaTemplate 创建角色模板 POST /api/admin/personaTemplates
func TestCreatePersonaTemplate(t *testing.T) {
	req := vo.AdminCreatePersonaTemplateReq{
		Name:        "编程助手",
		Icon:        "code",
		Description: "擅长代码编写、调试和架构设计",
		RoleInfo:    "你是一个专业的编程助手，擅长多种编程语言和技术栈。帮助用户编写高质量代码、调试问题并提供架构建议。",
		CategoryId:  7,
		SortOrder:   0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreatePersonaTemplate2 创建第二个角色模板
func TestCreatePersonaTemplate2(t *testing.T) {
	req := vo.AdminCreatePersonaTemplateReq{
		Name:        "翻译专家",
		Icon:        "languages",
		Description: "精通多语言翻译",
		RoleInfo:    "你是一个专业的翻译专家，精通中英日韩等多种语言。帮助用户进行准确的翻译，保持原文的语气和风格。",
		CategoryId:  1,
		SortOrder:   0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetPersonaTemplateDetail 角色模板详情 GET /api/admin/personaTemplates/:id
func TestGetPersonaTemplateDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdatePersonaTemplate 编辑角色模板 PUT /api/admin/personaTemplates/:id
func TestUpdatePersonaTemplate(t *testing.T) {
	req := vo.AdminUpdatePersonaTemplateReq{
		Name:        stringPtr("高级编程助手"),
		Icon:        stringPtr("code-2"),
		Description: stringPtr("擅长全栈开发和系统架构"),
		SortOrder:   intPtr(1),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdatePersonaTemplateRoleInfo 编辑角色模板的 roleInfo
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

// TestGetPersonaTemplateNotFound 查询不存在的角色模板
func TestGetPersonaTemplateNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreatePersonaTemplateMissingRoleInfo 创建角色模板缺少 roleInfo
func TestCreatePersonaTemplateMissingRoleInfo(t *testing.T) {
	req := vo.AdminCreatePersonaTemplateReq{
		Name:       "空提示词模板",
		Icon:       "bot",
		RoleInfo:   "",
		CategoryId: 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestDeletePersonaTemplate 删除角色模板 DELETE /api/admin/personaTemplates/:id
func TestDeletePersonaTemplate(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaTemplates/2").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// ════════════════════════════════════════════════════════════════════════════
// 角色分类相关测试
// ════════════════════════════════════════════════════════════════════════════

// TestGetPersonaCategories 角色分类列表 GET /api/admin/personaCategories
func TestGetPersonaCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetAllPersonaCategories 获取所有角色分类 GET /api/admin/personaCategories/all
func TestGetAllPersonaCategories(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/all").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetPersonaCategoryDetail 角色分类详情 GET /api/admin/personaCategories/:id
func TestGetPersonaCategoryDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreatePersonaCategory 创建角色分类 POST /api/admin/personaCategories
func TestCreatePersonaCategory(t *testing.T) {
	req := vo.AdminCreatePersonaCategoryReq{
		Name: "自定义分类",
		Icon: "sparkles",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdatePersonaCategory 编辑角色分类 PUT /api/admin/personaCategories/:id
func TestUpdatePersonaCategory(t *testing.T) {
	req := vo.AdminUpdatePersonaCategoryReq{
		Name: stringPtr("更新分类名"),
		Icon: stringPtr("star"),
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/7").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestDeletePersonaCategory 删除角色分类 DELETE /api/admin/personaCategories/:id
func TestDeletePersonaCategory(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/8").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetPersonaCategoryNotFound 查询不存在的角色分类
func TestGetPersonaCategoryNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/personaCategories/99999").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
