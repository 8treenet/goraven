package controller

import (
	"raven/backend/infra"
	"raven/backend/vo"
	"raven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/personas", &PersonaController{}, infra.NewAuth(true))
	})
}

type PersonaController struct {
	PersonaSev	*service.PersonaService
	McpSev		*service.McpService
	SkillSev	*service.SkillService
	Request		*infra.Request
}

func (controller *PersonaController) BeforeActivation(b freedom.BeforeActivation) {

	b.Handle("GET", "/", "GetPersonas")
	b.Handle("GET", "/{id:int}", "GetPersona")
	b.Handle("POST", "/", "CreatePersona")
	b.Handle("PUT", "/{id:int}", "UpdatePersona")
	b.Handle("DELETE", "/{id:int}", "DeletePersona")

	b.Handle("GET", "/personaTemplates", "GetPersonaTemplates")
	b.Handle("GET", "/personaTemplates/{id:int}", "GetPersonaTemplateDetail")

	b.Handle("GET", "/personaCategories", "GetPersonaCategories")
}

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

func (controller *PersonaController) GetPersona(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.PersonaSev.GetUserPersona(id, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

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

func (controller *PersonaController) DeletePersona(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.PersonaSev.DeleteUserPersona(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

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

func (controller *PersonaController) GetPersonaTemplateDetail(id int) freedom.Result {
	rsp, err := controller.PersonaSev.GetPersonaTemplateDetailForUser(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *PersonaController) GetPersonaCategories() freedom.Result {
	rsp, err := controller.PersonaSev.ListPersonaCategoriesForUser()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}
