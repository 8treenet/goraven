package controller

import (
	"goraven/backend/infra"
	"goraven/backend/vo"
	"goraven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/personas", &PersonaController{}, infra.NewAuth(true))
	})
}

// PersonaController 用户角色相关接口
type PersonaController struct {
	PersonaSev *service.PersonaService
	McpSev     *service.McpService
	SkillSev   *service.SkillService
	Request    *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/personas
func (controller *PersonaController) BeforeActivation(b freedom.BeforeActivation) {
	// 用户角色 CRUD
	b.Handle("GET", "/", "GetPersonas")
	b.Handle("GET", "/{id:int}", "GetPersona")
	b.Handle("POST", "/", "CreatePersona")
	b.Handle("PUT", "/{id:int}", "UpdatePersona")
	b.Handle("DELETE", "/{id:int}", "DeletePersona")
	// 角色模板（用户端只读）
	b.Handle("GET", "/personaTemplates", "GetPersonaTemplates")
	b.Handle("GET", "/personaTemplates/{id:int}", "GetPersonaTemplateDetail")
	// 角色分类（用户端只读）
	b.Handle("GET", "/personaCategories", "GetPersonaCategories")
}

// GetPersonas 获取当前用户角色列表 GET /api/personas?simple=true
func (controller *PersonaController) GetPersonas() freedom.Result {
	userId := controller.Request.GetUserId()

	simple := controller.Request.Worker().IrisContext().URLParam("simple") == "true"
	if simple {
		rsp, err := controller.PersonaSev.ListUserPersonasSimple(userId)
		if err != nil {
			return &infra.JSONResponse{Error: err}
		}
		return &infra.JSONResponse{Object: rsp}
	}

	rsp, err := controller.PersonaSev.ListUserPersonas(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetPersona 获取用户角色详情（编辑页回填） GET /api/personas/:id
func (controller *PersonaController) GetPersona(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.PersonaSev.GetUserPersona(id, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// CreatePersona 创建用户角色 POST /api/personas
func (controller *PersonaController) CreatePersona() freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.CreateUserPersonaReq{}
	if err := controller.Request.ReadJSON(req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.PersonaSev.CreateUserPersona(userId, req, controller.McpSev, controller.SkillSev); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UpdatePersona 编辑用户角色 PUT /api/personas/:id
func (controller *PersonaController) UpdatePersona(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	req := &vo.UpdateUserPersonaReq{}
	if err := controller.Request.ReadJSON(req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.PersonaSev.UpdateUserPersona(id, userId, req, controller.McpSev, controller.SkillSev); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// DeletePersona 删除用户角色（软删除，关联 session.personaId 置零） DELETE /api/personas/:id
func (controller *PersonaController) DeletePersona(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.PersonaSev.DeleteUserPersona(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetPersonaTemplates 获取角色模板列表（支持按分类筛选） GET /api/personas/personaTemplates
func (controller *PersonaController) GetPersonaTemplates() freedom.Result {
	req := &vo.AdminPersonaTemplateListReq{}
	if err := controller.Request.ReadQuery(req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.PersonaSev.ListPersonaTemplatesForUser(req.CategoryId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetPersonaTemplateDetail 获取角色模板详情（含完整 roleInfo，用于预填角色设定） GET /api/personas/personaTemplates/:id
func (controller *PersonaController) GetPersonaTemplateDetail(id int) freedom.Result {
	rsp, err := controller.PersonaSev.GetPersonaTemplateDetailForUser(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetPersonaCategories 获取角色分类列表 GET /api/personas/personaCategories
func (controller *PersonaController) GetPersonaCategories() freedom.Result {
	rsp, err := controller.PersonaSev.ListPersonaCategoriesForUser()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}
