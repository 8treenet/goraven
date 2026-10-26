package controller

import (
	"strconv"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository/seed/mock"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/config"

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

// AdminController 管理后台控制器
// 路由前缀 /api/admin，仅 role=1 的管理员可访问
type AdminController struct {
	UserSev      *service.UserService          // 用户服务
	ModelSev     *service.AIModelService       // 模型服务
	McpSev       *service.McpService           // MCP 服务
	SkillSev     *service.SkillService         // 技能服务
	PersonaSev   *service.PersonaService       // 角色模板服务
	SystemSev    *service.SystemInfoService    // 系统信息服务
	DashboardSev *service.DashboardService     // 仪表盘服务
	SettingSev   *service.SystemSettingService // 系统设置服务
	TPSev        *service.TeamProjectService   // 团队项目服务
	Request      *infra.Request                // 请求工具
	Worker       freedom.Worker                // 工作空间
}

// BeforeActivation 绑定管理后台路由
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
	b.Handle("PUT", "/models/{id:string}/flash", "SetFlashModel")
	b.Handle("PUT", "/models/{id:string}/visual", "SetVisualModel")
	b.Handle("GET", "/models/{id:string}/members", "ListModelMembers")
	b.Handle("PUT", "/models/{id:string}/members", "UpdateModelMembers")
	b.Handle("PUT", "/models/{id:string}/access", "UpdateModelAccess")

	b.Handle("GET", "/providers", "GetProviders")
	b.Handle("GET", "/providers/recommend", "GetRecommendModels")

	b.Handle("GET", "/mcp", "GetMCPs")
	b.Handle("GET", "/mcp/{id:int}", "GetMCPDetail")
	b.Handle("POST", "/mcp", "CreateMCP")
	b.Handle("PUT", "/mcp/{id:int}", "UpdateMCP")
	b.Handle("DELETE", "/mcp/{id:int}", "DeleteMCP")
	b.Handle("PUT", "/mcp/{id:int}/status", "UpdateMCPStatus")
	b.Handle("PUT", "/mcp/{id:int}/alwaysOn", "ToggleMCPAlwaysOn")
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

	b.Handle("GET", "/sharedProjects", "GetSharedProjects")
	b.Handle("PUT", "/sharedProjects/{id:int}", "UpdateSharedProject")
	b.Handle("DELETE", "/sharedProjects/{id:int}", "UnshareSharedProject")

	b.Handle("GET", "/dashboard", "GetDashboard")
	b.Handle("GET", "/dashboard/tokenTrend", "GetTokenTrend")
	b.Handle("GET", "/dashboard/modelUsage", "GetModelUsage")
	b.Handle("GET", "/dashboard/userTokenRank", "GetUserTokenRank")
	b.Handle("GET", "/dashboard/activeUsers", "GetActiveUsers")
}

// GetUsers 用户列表 GET /api/admin/users
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

// CreateUser 创建用户 POST /api/admin/users
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

// BatchGetUsers 批量查询用户 POST /api/admin/users/batch
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

// GetUserDetail 用户详情 GET /api/admin/users/:userId
func (controller *AdminController) GetUserDetail(userId string) freedom.Result {
	detail, err := controller.UserSev.AdminGetUserDetail(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// UpdateUser 编辑用户 PUT /api/admin/users/:userId
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

// ResetPassword 重置密码 PUT /api/admin/users/:userId/reset-password
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

// DeleteUser 删除用户 DELETE /api/admin/users/:userId
func (controller *AdminController) DeleteUser(userId string) freedom.Result {
	if err := controller.UserSev.AdminDeleteUser(userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetModels 模型列表 GET /api/admin/models
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

// CreateModel 创建模型 POST /api/admin/models
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

// UpdateModel 编辑模型 PUT /api/admin/models/:id
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

// DeleteModel 删除模型 DELETE /api/admin/models/:id
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

// UpdateModelStatus 启用/禁用模型 PUT /api/admin/models/:id/status
func (controller *AdminController) UpdateModelStatus(id string) freedom.Result {
	var req struct {
		Status uint8 `json:"status"`
	}
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	/*
		modelId, err := strconv.Atoi(id)
		if err != nil {
			return &infra.JSONResponse{Error: err}
		}
		if err := controller.ModelSev.UpdateModelStatus(modelId, req.Status); err != nil {
			return &infra.JSONResponse{Error: err}
		}
	*/

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetModelDetail 模型详情 GET /api/admin/models/:id
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

// SetDefaultModel 切换默认模型开关（加入/移出默认池） PUT /api/admin/models/:id/default
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

// SetFlashModel 设为 Flash 模型 PUT /api/admin/models/:id/flash
func (controller *AdminController) SetFlashModel(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.SetFlashModel(modelId); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// SetVisualModel 设为多模态识别模型 PUT /api/admin/models/:id/visual
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

// ListModelMembers 模型成员列表 GET /api/admin/models/:id/members
func (controller *AdminController) ListModelMembers(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.ModelSev.ListMembers(modelId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// UpdateModelMembers 编辑模型成员 PUT /api/admin/models/:id/members
func (controller *AdminController) UpdateModelMembers(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	var req vo.AIModelMemberUpdateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.UpdateMembers(modelId, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UpdateModelAccess 设置模型访问权限 PUT /api/admin/models/:id/access
func (controller *AdminController) UpdateModelAccess(id string) freedom.Result {
	modelId, err := strconv.Atoi(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	var req vo.AIModelAccessUpdateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.ModelSev.UpdateAccess(modelId, req.Access); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetProviders 供应商列表 GET /api/admin/providers
func (controller *AdminController) GetProviders() freedom.Result {
	list := controller.ModelSev.ListProviders()
	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

// GetMCPs MCP 列表 GET /api/admin/mcp
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

// GetMCPDetail MCP 详情 GET /api/admin/mcp/:id
func (controller *AdminController) GetMCPDetail(id int) freedom.Result {
	detail, err := controller.McpSev.GetMCPEndpointDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// CreateMCP 创建 MCP POST /api/admin/mcp
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

// UpdateMCP 编辑 MCP PUT /api/admin/mcp/:id
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

// DeleteMCP 删除 MCP DELETE /api/admin/mcp/:id
func (controller *AdminController) DeleteMCP(id int) freedom.Result {
	if err := controller.McpSev.DeleteMCPEndpoint(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UpdateMCPStatus 启用/禁用 MCP PUT /api/admin/mcp/:id/status
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

// ToggleMCPAlwaysOn 切换 MCP 始终启用 PUT /api/admin/mcp/:id/alwaysOn
func (controller *AdminController) ToggleMCPAlwaysOn(id int) freedom.Result {
	var req vo.AdminMCPToggleAlwaysOnReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.McpSev.ToggleMCPAlwaysOn(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetRecommendMCPs 推荐 MCP 列表 GET /api/admin/mcp/recommend
func (controller *AdminController) GetRecommendMCPs() freedom.Result {
	list := controller.McpSev.GetRecommendMCPs()
	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

// CheckMCPHealth 手动触发所有启用 MCP 的健康检查 POST /api/admin/mcp/healthCheck
func (controller *AdminController) CheckMCPHealth() freedom.Result {
	controller.McpSev.CheckAllMCPHealth()
	return &infra.JSONResponse{Object: map[string]string{"status": "checking"}}
}

// GetRecommendModels 推荐模型 GET /api/admin/providers/recommend
// 永远返回空列表不报错，用于前端选型辅助
func (controller *AdminController) GetRecommendModels() freedom.Result {
	var req struct {
		ProviderID string `url:"providerId"`
		APIKey     string `url:"apiKey"`
		BaseURL    string `url:"baseUrl"`
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

// GetSystemSkills 系统技能列表 GET /api/admin/systemSkills
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

// GetSystemSkillDetail 系统技能详情 GET /api/admin/systemSkills/:id
func (controller *AdminController) GetSystemSkillDetail(id int) freedom.Result {
	detail, err := controller.SkillSev.GetSystemSkillDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// CreateSystemSkill 创建系统技能 POST /api/admin/systemSkills
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

// UpdateSystemSkill 编辑系统技能 PUT /api/admin/systemSkills/:id
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

// UpdateSystemSkillStatus 启用/禁用系统技能 PUT /api/admin/systemSkills/:id/status
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

// DeleteSystemSkill 删除系统技能 DELETE /api/admin/systemSkills/:id
func (controller *AdminController) DeleteSystemSkill(id int) freedom.Result {
	if err := controller.SkillSev.DeleteSystemSkill(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetMarketSkills 市场技能列表 GET /api/admin/marketSkills
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

// GetMarketSkillDetail 市场技能详情 GET /api/admin/marketSkills/:id
func (controller *AdminController) GetMarketSkillDetail(id int) freedom.Result {
	detail, err := controller.SkillSev.GetMarketSkillDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// UpdateMarketSkill 编辑市场技能 PUT /api/admin/marketSkills/:id
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

// UpdateMarketSkillStatus 上架/下架市场技能 PUT /api/admin/marketSkills/:id/status
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

// DeleteMarketSkill 删除市场技能 DELETE /api/admin/marketSkills/:id?cascade=true
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

// GetMarketSkillUsers 市场技能已安装用户列表 GET /api/admin/marketSkills/:id/users
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

// PublishMarketSkill 发布市场技能 POST /api/admin/marketSkills/publish
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

// ImportClawHubSkill 从 ClawHub 导入技能 POST /api/admin/marketSkills/import
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

// SearchClawHub 搜索 ClawHub 技能 GET /api/admin/clawhub/search
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

// ExploreClawHub 浏览 ClawHub 技能列表 GET /api/admin/clawhub/explore
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

// GetClawHubSkillDetail ClawHub 技能详情 GET /api/admin/clawhub/skills/:slug
func (controller *AdminController) GetClawHubSkillDetail(slug string) freedom.Result {
	detail, err := controller.SkillSev.GetClawHubSkillDetail(slug)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// GetSkillCategories 技能分类列表 GET /api/admin/skillCategories
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

// GetAllSkillCategories 获取所有分类 GET /api/admin/skillCategories/all
func (controller *AdminController) GetAllSkillCategories() freedom.Result {
	list, err := controller.SkillSev.GetAllSkillCategories()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

// GetSkillCategoryDetail 技能分类详情 GET /api/admin/skillCategories/:id
func (controller *AdminController) GetSkillCategoryDetail(id int) freedom.Result {
	detail, err := controller.SkillSev.GetSkillCategoryDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// CreateSkillCategory 创建技能分类 POST /api/admin/skillCategories
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

// UpdateSkillCategory 编辑技能分类 PUT /api/admin/skillCategories/:id
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

// DeleteSkillCategory 删除技能分类 DELETE /api/admin/skillCategories/:id
func (controller *AdminController) DeleteSkillCategory(id int) freedom.Result {
	if err := controller.SkillSev.DeleteSkillCategory(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ════════════════════════════════════════════════════════════════════════════
// 角色模板
// ════════════════════════════════════════════════════════════════════════════

// GetPersonaTemplates 角色模板列表 GET /api/admin/personaTemplates
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

// GetPersonaTemplateDetail 角色模板详情 GET /api/admin/personaTemplates/:id
func (controller *AdminController) GetPersonaTemplateDetail(id int) freedom.Result {
	detail, err := controller.PersonaSev.GetPersonaTemplateDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// CreatePersonaTemplate 创建角色模板 POST /api/admin/personaTemplates
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

// UpdatePersonaTemplate 编辑角色模板 PUT /api/admin/personaTemplates/:id
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

// DeletePersonaTemplate 删除角色模板 DELETE /api/admin/personaTemplates/:id
func (controller *AdminController) DeletePersonaTemplate(id int) freedom.Result {
	if err := controller.PersonaSev.DeletePersonaTemplate(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ════════════════════════════════════════════════════════════════════════════
// 角色分类
// ════════════════════════════════════════════════════════════════════════════

// GetPersonaCategories 角色分类列表 GET /api/admin/personaCategories
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

// GetAllPersonaCategories 获取所有角色分类 GET /api/admin/personaCategories/all
func (controller *AdminController) GetAllPersonaCategories() freedom.Result {
	list, err := controller.PersonaSev.GetAllPersonaCategories()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"list": list}}
}

// GetPersonaCategoryDetail 角色分类详情 GET /api/admin/personaCategories/:id
func (controller *AdminController) GetPersonaCategoryDetail(id int) freedom.Result {
	detail, err := controller.PersonaSev.GetPersonaCategoryDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: detail}
}

// CreatePersonaCategory 创建角色分类 POST /api/admin/personaCategories
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

// UpdatePersonaCategory 编辑角色分类 PUT /api/admin/personaCategories/:id
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

// DeletePersonaCategory 删除角色分类 DELETE /api/admin/personaCategories/:id
func (controller *AdminController) DeletePersonaCategory(id int) freedom.Result {
	if err := controller.PersonaSev.DeletePersonaCategory(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GetSystemInfo 系统信息 GET /api/admin/systemInfo?forceRefresh=true
func (controller *AdminController) GetSystemInfo() freedom.Result {
	forceRefresh := controller.Worker.IrisContext().URLParam("forceRefresh") == "true"

	rsp, err := controller.SystemSev.GetSystemInfo(forceRefresh)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// GetDashboard 管理员仪表盘聚合数据 GET /api/admin/dashboard
func (controller *AdminController) GetDashboard() freedom.Result {
	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAdminDashboard()}
	}

	rsp, err := controller.DashboardSev.GetAdminDashboard()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// GetTokenTrend 管理员全局 Token 趋势 GET /api/admin/dashboard/tokenTrend?days=30
func (controller *AdminController) GetTokenTrend() freedom.Result {
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAdminTokenTrend(req.Days)}
	}

	rsp, err := controller.DashboardSev.GetAdminTokenTrend(req.Days)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// GetModelUsage 管理员全局模型使用分布 GET /api/admin/dashboard/modelUsage?days=30
func (controller *AdminController) GetModelUsage() freedom.Result {
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAdminModelUsage(req.Days)}
	}

	rsp, err := controller.DashboardSev.GetAdminModelUsage(req.Days)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// GetUserTokenRank 管理员用户 Token 消耗排行 GET /api/admin/dashboard/userTokenRank?days=30
func (controller *AdminController) GetUserTokenRank() freedom.Result {
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAdminUserTokenRank(req.Days)}
	}

	rsp, err := controller.DashboardSev.GetAdminUserTokenRank(req.Days)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// GetActiveUsers 管理员全局活跃用户趋势 GET /api/admin/dashboard/activeUsers?days=30
func (controller *AdminController) GetActiveUsers() freedom.Result {
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAdminActiveUserTrend(req.Days)}
	}

	rsp, err := controller.DashboardSev.GetAdminActiveUserTrend(req.Days)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// GetSettings 获取全部系统设置（含 UI 元数据） GET /api/admin/settings
func (controller *AdminController) GetSettings() freedom.Result {
	groups, err := controller.SettingSev.GetSettings()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]interface{}{"groups": groups}}
}

// UpdateSettings 批量更新系统设置 PUT /api/admin/settings
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

// GetSharedProjects 团队项目列表 GET /api/admin/sharedProjects
func (controller *AdminController) GetSharedProjects() freedom.Result {
	var req vo.AdminTeamProjectListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.TPSev.AdminListTeamProjects(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// UpdateSharedProject 管理端编辑团队项目简介 PUT /api/admin/sharedProjects/:id
func (controller *AdminController) UpdateSharedProject(id int) freedom.Result {
	var req vo.TeamProjectUpdateReq
	if err := controller.Request.ReadJSON(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.AdminUpdateDescription(id, req.Description); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UnshareSharedProject 管理端删除团队项目 DELETE /api/admin/sharedProjects/:id
func (controller *AdminController) UnshareSharedProject(id int) freedom.Result {
	if err := controller.TPSev.AdminDeleteProject(id); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}
