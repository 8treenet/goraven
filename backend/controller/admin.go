package controller

import (
	"strconv"

	"raven/backend/infra"
	"raven/backend/po"
	"raven/backend/service"
	"raven/backend/vo"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/admin", &AdminController{}, infra.NewAuth(true), func(ctx freedom.Context) {
			worker := freedom.ToWorker(ctx)
			worker.Store().Get(infra.UserRoleStoreKey)
			v := worker.Store().Get(infra.UserRoleStoreKey)
			if v == nil {
				ctx.JSON(infra.ResBodyObject{Code: infra.Access_Denied, Msg: "Permission denied"})
				return
			}
			if role, ok := v.(uint8); ok && role == po.UserRoleAdmin {
				ctx.Next()
				return
			}
			ctx.JSON(infra.ResBodyObject{Code: infra.Access_Denied, Msg: "Permission denied"})
		})
	})
}

type AdminController struct {
	UserSev		*service.UserService
	ModelSev	*service.AIModelService
	McpSev		*service.McpService
	SkillSev	*service.SkillService
	PersonaSev	*service.PersonaService
	SystemSev	*service.SystemInfoService
	DashboardSev	*service.DashboardService
	SettingSev	*service.SystemSettingService
	Request		*infra.Request
	Worker		freedom.Worker
}

func (controller *AdminController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/users", "GetUsers")
	b.Handle("POST", "/users", "CreateUser")
	b.Handle("POST", "/users/batch", "BatchGetUsers")
	b.Handle("GET", "/users/{userId:string}", "GetUserDetail")
	b.Handle("PUT", "/users/{userId:string}", "UpdateUser")
	b.Handle("PUT", "/users/{userId:string}/reset-password", "ResetPassword")
	b.Handle("DELETE", "/users/{userId:string}", "DeleteUser")

	b.Handle("GET", "/models", "GetModels")
	b.Handle("POST", "/models", "CreateModel")
	b.Handle("PUT", "/models/{id:string}", "UpdateModel")
	b.Handle("DELETE", "/models/{id:string}", "DeleteModel")
	b.Handle("GET", "/models/{id:string}", "GetModelDetail")
	b.Handle("PUT", "/models/{id:string}/status", "UpdateModelStatus")
	b.Handle("PUT", "/models/{id:string}/default", "SetDefaultModel")
	b.Handle("PUT", "/models/{id:string}/compress", "SetCompressModel")
	b.Handle("PUT", "/models/{id:string}/visual", "SetVisualModel")

	b.Handle("GET", "/providers", "GetProviders")
	b.Handle("GET", "/providers/recommend", "GetRecommendModels")

	b.Handle("GET", "/mcp", "GetMCPs")
	b.Handle("GET", "/mcp/{id:int}", "GetMCPDetail")
	b.Handle("POST", "/mcp", "CreateMCP")
	b.Handle("PUT", "/mcp/{id:int}", "UpdateMCP")
	b.Handle("DELETE", "/mcp/{id:int}", "DeleteMCP")
	b.Handle("PUT", "/mcp/{id:int}/status", "UpdateMCPStatus")
	b.Handle("GET", "/mcp/recommend", "GetRecommendMCPs")
	b.Handle("POST", "/mcp/healthCheck", "CheckMCPHealth")

	b.Handle("GET", "/systemSkills", "GetSystemSkills")
	b.Handle("GET", "/systemSkills/{id:int}", "GetSystemSkillDetail")
	b.Handle("POST", "/systemSkills", "CreateSystemSkill")
	b.Handle("PUT", "/systemSkills/{id:int}", "UpdateSystemSkill")
	b.Handle("PUT", "/systemSkills/{id:int}/status", "UpdateSystemSkillStatus")
	b.Handle("DELETE", "/systemSkills/{id:int}", "DeleteSystemSkill")

	b.Handle("GET", "/marketSkills", "GetMarketSkills")
	b.Handle("GET", "/marketSkills/{id:int}", "GetMarketSkillDetail")
	b.Handle("PUT", "/marketSkills/{id:int}", "UpdateMarketSkill")
	b.Handle("PUT", "/marketSkills/{id:int}/status", "UpdateMarketSkillStatus")
	b.Handle("DELETE", "/marketSkills/{id:int}", "DeleteMarketSkill")
	b.Handle("GET", "/marketSkills/{id:int}/users", "GetMarketSkillUsers")
	b.Handle("POST", "/marketSkills/publish", "PublishMarketSkill")
	b.Handle("POST", "/marketSkills/import", "ImportClawHubSkill")

	b.Handle("GET", "/clawhub/search", "SearchClawHub")
	b.Handle("GET", "/clawhub/explore", "ExploreClawHub")
	b.Handle("GET", "/clawhub/skills/{slug:string}", "GetClawHubSkillDetail")

	b.Handle("GET", "/skillCategories", "GetSkillCategories")
	b.Handle("GET", "/skillCategories/all", "GetAllSkillCategories")
	b.Handle("GET", "/skillCategories/{id:int}", "GetSkillCategoryDetail")
	b.Handle("POST", "/skillCategories", "CreateSkillCategory")
	b.Handle("PUT", "/skillCategories/{id:int}", "UpdateSkillCategory")
	b.Handle("DELETE", "/skillCategories/{id:int}", "DeleteSkillCategory")

	b.Handle("GET", "/personaTemplates", "GetPersonaTemplates")
	b.Handle("GET", "/personaTemplates/{id:int}", "GetPersonaTemplateDetail")
	b.Handle("POST", "/personaTemplates", "CreatePersonaTemplate")
	b.Handle("PUT", "/personaTemplates/{id:int}", "UpdatePersonaTemplate")
	b.Handle("DELETE", "/personaTemplates/{id:int}", "DeletePersonaTemplate")

	b.Handle("GET", "/personaCategories", "GetPersonaCategories")
	b.Handle("GET", "/personaCategories/all", "GetAllPersonaCategories")
	b.Handle("GET", "/personaCategories/{id:int}", "GetPersonaCategoryDetail")
	b.Handle("POST", "/personaCategories", "CreatePersonaCategory")
	b.Handle("PUT", "/personaCategories/{id:int}", "UpdatePersonaCategory")
	b.Handle("DELETE", "/personaCategories/{id:int}", "DeletePersonaCategory")

	b.Handle("GET", "/systemInfo", "GetSystemInfo")

	b.Handle("GET", "/settings", "GetSettings")
	b.Handle("PUT", "/settings", "UpdateSettings")

	b.Handle("GET", "/dashboard", "GetDashboard")
	b.Handle("GET", "/dashboard/tokenTrend", "GetTokenTrend")
	b.Handle("GET", "/dashboard/activeUsers", "GetActiveUsers")
}

func (controller *AdminController) GetUsers() freedom.Result {
	var req vo.AdminUserListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.UserSev.AdminListUsers(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) CreateUser() freedom.Result {
	var req vo.AdminCreateUserReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.UserSev.AdminCreateUser(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) BatchGetUsers() freedom.Result {
	var req vo.AdminBatchUserReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	items, err := controller.UserSev.AdminBatchGetUsers(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"list": items}}
}

func (controller *AdminController) GetUserDetail(userId string) freedom.Result {
	detail, err := controller.UserSev.AdminGetUserDetail(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) UpdateUser(userId string) freedom.Result {
	var req vo.AdminUpdateUserReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.UserSev.AdminUpdateUser(userId, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) ResetPassword(userId string) freedom.Result {
	var req vo.AdminResetPasswordReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.UserSev.AdminResetPassword(userId, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeleteUser(userId string) freedom.Result {
	if err := controller.UserSev.AdminDeleteUser(userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetModels() freedom.Result {
	var req vo.AdminModelListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.ModelSev.ListModels(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) CreateModel() freedom.Result {
	var req vo.AdminCreateModelReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.CreateModel(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateModel(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	var req vo.AdminUpdateModelReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.UpdateModel(modelId, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeleteModel(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.DeleteModel(modelId); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateModelStatus(id string) freedom.Result {
	var req struct {
		Status uint8 `json:"status"`
	}
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetModelDetail(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	detail, err := controller.ModelSev.GetModelDetail(modelId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) SetDefaultModel(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.SetDefaultModel(modelId); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) SetCompressModel(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.SetCompressModel(modelId); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) SetVisualModel(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.SetVisualModel(modelId); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetProviders() freedom.Result {
	list := controller.ModelSev.ListProviders()
	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

func (controller *AdminController) GetMCPs() freedom.Result {
	var req vo.AdminMCPListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.McpSev.ListMCPEndpoints(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetMCPDetail(id int) freedom.Result {
	detail, err := controller.McpSev.GetMCPEndpointDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) CreateMCP() freedom.Result {
	var req vo.AdminCreateMCPReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.McpSev.CreateMCPEndpoint(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateMCP(id int) freedom.Result {
	var req vo.AdminUpdateMCPReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.McpSev.UpdateMCPEndpoint(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeleteMCP(id int) freedom.Result {
	if err := controller.McpSev.DeleteMCPEndpoint(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateMCPStatus(id int) freedom.Result {
	var req struct {
		Status uint8 `json:"status"`
	}
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.McpSev.UpdateMCPEndpointStatus(id, req.Status); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetRecommendMCPs() freedom.Result {
	list := controller.McpSev.GetRecommendMCPs()
	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

func (controller *AdminController) CheckMCPHealth() freedom.Result {
	controller.McpSev.CheckAllMCPHealth()
	return &infra.JSONResponse{Object: map[string]string{"status": "checking"}}
}

func (controller *AdminController) GetRecommendModels() freedom.Result {
	var req struct {
		ProviderID	string	`url:"providerId"`
		APIKey		string	`url:"apiKey"`
		BaseURL		string	`url:"baseUrl"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Object: map[string]interface{}{"list": []interface{}{}}}
	}

	list, _ := controller.ModelSev.RecommendModels(req.ProviderID, req.APIKey, req.BaseURL)
	if list == nil {
		list = []vo.RecommendModelItem{}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

func (controller *AdminController) GetSystemSkills() freedom.Result {
	var req vo.AdminSystemSkillListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.SkillSev.ListSystemSkills(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetSystemSkillDetail(id int) freedom.Result {
	detail, err := controller.SkillSev.GetSystemSkillDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) CreateSystemSkill() freedom.Result {
	var req vo.AdminCreateSystemSkillReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.CreateSystemSkill(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateSystemSkill(id int) freedom.Result {
	var req vo.AdminUpdateSystemSkillReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.UpdateSystemSkill(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateSystemSkillStatus(id int) freedom.Result {
	var req vo.AdminSystemSkillStatusReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.UpdateSystemSkillStatus(id, *req.Status); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeleteSystemSkill(id int) freedom.Result {
	if err := controller.SkillSev.DeleteSystemSkill(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetMarketSkills() freedom.Result {
	var req vo.AdminMarketSkillListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.SkillSev.ListMarketSkills(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetMarketSkillDetail(id int) freedom.Result {
	detail, err := controller.SkillSev.GetMarketSkillDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) UpdateMarketSkill(id int) freedom.Result {
	var req vo.AdminUpdateMarketSkillReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.UpdateMarketSkill(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateMarketSkillStatus(id int) freedom.Result {
	var req vo.AdminMarketSkillStatusReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.UpdateMarketSkillStatus(id, *req.Status); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeleteMarketSkill(id int) freedom.Result {
	var req vo.AdminDeleteMarketSkillReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.DeleteMarketSkill(id, req.Cascade); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetMarketSkillUsers(id int) freedom.Result {
	var req vo.AdminMarketSkillUserListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.SkillSev.GetMarketSkillUsers(id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) PublishMarketSkill() freedom.Result {
	var req vo.AdminPublishMarketSkillReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.PublishMarketSkill(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) ImportClawHubSkill() freedom.Result {
	var req vo.AdminImportClawHubSkillReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.ImportClawHubSkill(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) SearchClawHub() freedom.Result {
	var req vo.ClawHubSearchReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.SkillSev.SearchClawHub(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) ExploreClawHub() freedom.Result {
	var req vo.ClawHubExploreReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.SkillSev.ExploreClawHub(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetClawHubSkillDetail(slug string) freedom.Result {
	detail, err := controller.SkillSev.GetClawHubSkillDetail(slug)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) GetSkillCategories() freedom.Result {
	var req vo.AdminSkillCategoryListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.SkillSev.ListSkillCategories(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetAllSkillCategories() freedom.Result {
	list, err := controller.SkillSev.GetAllSkillCategories()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

func (controller *AdminController) GetSkillCategoryDetail(id int) freedom.Result {
	detail, err := controller.SkillSev.GetSkillCategoryDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) CreateSkillCategory() freedom.Result {
	var req vo.AdminCreateSkillCategoryReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.CreateSkillCategory(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdateSkillCategory(id int) freedom.Result {
	var req vo.AdminUpdateSkillCategoryReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.SkillSev.UpdateSkillCategory(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeleteSkillCategory(id int) freedom.Result {
	if err := controller.SkillSev.DeleteSkillCategory(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetPersonaTemplates() freedom.Result {
	var req vo.AdminPersonaTemplateListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.PersonaSev.ListPersonaTemplates(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetPersonaTemplateDetail(id int) freedom.Result {
	detail, err := controller.PersonaSev.GetPersonaTemplateDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) CreatePersonaTemplate() freedom.Result {
	var req vo.AdminCreatePersonaTemplateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.PersonaSev.CreatePersonaTemplate(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdatePersonaTemplate(id int) freedom.Result {
	var req vo.AdminUpdatePersonaTemplateReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.PersonaSev.UpdatePersonaTemplate(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeletePersonaTemplate(id int) freedom.Result {
	if err := controller.PersonaSev.DeletePersonaTemplate(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetPersonaCategories() freedom.Result {
	var req vo.AdminPersonaCategoryListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.PersonaSev.ListPersonaCategories(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetAllPersonaCategories() freedom.Result {
	list, err := controller.PersonaSev.GetAllPersonaCategories()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

func (controller *AdminController) GetPersonaCategoryDetail(id int) freedom.Result {
	detail, err := controller.PersonaSev.GetPersonaCategoryDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

func (controller *AdminController) CreatePersonaCategory() freedom.Result {
	var req vo.AdminCreatePersonaCategoryReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.PersonaSev.CreatePersonaCategory(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) UpdatePersonaCategory(id int) freedom.Result {
	var req vo.AdminUpdatePersonaCategoryReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.PersonaSev.UpdatePersonaCategory(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) DeletePersonaCategory(id int) freedom.Result {
	if err := controller.PersonaSev.DeletePersonaCategory(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *AdminController) GetSystemInfo() freedom.Result {
	forceRefresh := controller.Worker.IrisContext().URLParam("forceRefresh") == "true"

	rsp, err := controller.SystemSev.GetSystemInfo(forceRefresh)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetDashboard() freedom.Result {
	refresh := controller.Worker.IrisContext().URLParam("refresh") == "true"
	rsp, err := controller.DashboardSev.GetAdminDashboard(refresh)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetTokenTrend() freedom.Result {
	refresh := controller.Worker.IrisContext().URLParam("refresh") == "true"
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	rsp, err := controller.DashboardSev.GetAdminTokenTrend(req.Days, refresh)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetActiveUsers() freedom.Result {
	refresh := controller.Worker.IrisContext().URLParam("refresh") == "true"
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	rsp, err := controller.DashboardSev.GetAdminActiveUserTrend(req.Days, refresh)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *AdminController) GetSettings() freedom.Result {
	groups, err := controller.SettingSev.GetSettings()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"groups": groups}}
}

func (controller *AdminController) UpdateSettings() freedom.Result {
	var req vo.AdminUpdateSettingsReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.SettingSev.UpdateSettings(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}
