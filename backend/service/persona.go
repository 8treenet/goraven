package service

import (
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/util"
	"unicode/utf8"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *PersonaService {
			return &PersonaService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *PersonaService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// PersonaService 角色模板与分类服务
type PersonaService struct {
	Worker      freedom.Worker
	PersonaRepo *repository.PersonaRepository
	ModelRepo   *repository.ProviderRepository
	McpRepo     *repository.MCPRepository
	SkillRepo   *repository.SkillRepository
}

// ════════════════════════════════════════════════════════════════════════════
// 角色模板
// ════════════════════════════════════════════════════════════════════════════

// ListPersonaTemplates 角色模板分页列表
func (service *PersonaService) ListPersonaTemplates(req *vo.AdminPersonaTemplateListReq) (*infra.PageResponse, error) {
	templates, pr, err := service.PersonaRepo.PaginatePersonaTemplates(req)
	if err != nil {
		return nil, err
	}

	categoryMap := service.batchGetCategoryMap(templates)

	items := make([]vo.AdminPersonaTemplateItem, 0, len(templates))
	for _, t := range templates {
		item := vo.AdminPersonaTemplateItem{
			TemplateId:  t.TemplateId,
			Name:        t.Name,
			Icon:        t.Icon,
			Description: t.Description,
			RoleInfo:    util.TruncateRunes(t.RoleInfo, 50),
			CategoryId:  t.CategoryId,
			UsageCount:  t.UsageCount,
			SortOrder:   t.SortOrder,
			Updated:     t.Updated,
		}
		if cat, ok := categoryMap[t.CategoryId]; ok {
			item.CategoryName = cat.Name
			item.CategoryIcon = cat.Icon
		}
		items = append(items, item)
	}

	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

// GetPersonaTemplateDetail 角色模板详情（含完整 roleInfo，用于编辑回填）
func (service *PersonaService) GetPersonaTemplateDetail(templateId int) (*vo.AdminPersonaTemplateDetailRsp, error) {
	tmpl, err := service.PersonaRepo.GetPersonaTemplateByID(templateId)
	if err != nil {
		return nil, errs.ErrPersonaTemplateNotFound
	}

	rsp := &vo.AdminPersonaTemplateDetailRsp{
		TemplateId:  tmpl.TemplateId,
		Name:        tmpl.Name,
		Icon:        tmpl.Icon,
		Description: tmpl.Description,
		RoleInfo:    tmpl.RoleInfo,
		CategoryId:  tmpl.CategoryId,
		UsageCount:  tmpl.UsageCount,
		SortOrder:   tmpl.SortOrder,
		Created:     tmpl.Created,
		Updated:     tmpl.Updated,
	}

	if tmpl.CategoryId > 0 {
		if cat, err := service.PersonaRepo.GetPersonaCategoryByID(tmpl.CategoryId); err == nil {
			rsp.CategoryName = cat.Name
			rsp.CategoryIcon = cat.Icon
		}
	}

	return rsp, nil
}

// CreatePersonaTemplate 创建角色模板
func (service *PersonaService) CreatePersonaTemplate(req *vo.AdminCreatePersonaTemplateReq) error {
	if req.RoleInfo == "" {
		return errs.ErrPersonaTemplateRoleInfoRequired
	}
	if utf8.RuneCountInString(req.RoleInfo) > 500 {
		return errs.ErrPersonaTemplateRoleInfoTooLong
	}
	if _, err := service.PersonaRepo.GetPersonaCategoryByID(req.CategoryId); err != nil {
		return errs.ErrPersonaCategoryNotFound
	}

	tmpl := &po.PersonaTemplate{
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
		RoleInfo:    req.RoleInfo,
		CategoryId:  req.CategoryId,
		SortOrder:   req.SortOrder,
	}
	return service.PersonaRepo.CreatePersonaTemplate(tmpl)
}

// UpdatePersonaTemplate 编辑角色模板
func (service *PersonaService) UpdatePersonaTemplate(templateId int, req *vo.AdminUpdatePersonaTemplateReq) error {
	if _, err := service.PersonaRepo.GetPersonaTemplateByID(templateId); err != nil {
		return errs.ErrPersonaTemplateNotFound
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.RoleInfo != nil {
		if *req.RoleInfo == "" {
			return errs.ErrPersonaTemplateRoleInfoRequired
		}
		if utf8.RuneCountInString(*req.RoleInfo) > 500 {
			return errs.ErrPersonaTemplateRoleInfoTooLong
		}
		updates["role_info"] = *req.RoleInfo
	}
	if req.CategoryId != nil {
		if _, err := service.PersonaRepo.GetPersonaCategoryByID(*req.CategoryId); err != nil {
			return errs.ErrPersonaCategoryNotFound
		}
		updates["category_id"] = *req.CategoryId
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if len(updates) == 0 {
		return nil
	}

	return service.PersonaRepo.UpdatePersonaTemplate(templateId, updates)
}

// DeletePersonaTemplate 删除角色模板（软删除）
func (service *PersonaService) DeletePersonaTemplate(templateId int) error {
	if _, err := service.PersonaRepo.GetPersonaTemplateByID(templateId); err != nil {
		return errs.ErrPersonaTemplateNotFound
	}
	return service.PersonaRepo.SoftDeletePersonaTemplate(templateId)
}

// ════════════════════════════════════════════════════════════════════════════
// 角色分类
// ════════════════════════════════════════════════════════════════════════════

// ListPersonaCategories 角色分类分页列表
func (service *PersonaService) ListPersonaCategories(req *vo.AdminPersonaCategoryListReq) (*infra.PageResponse, error) {
	categories, pr, err := service.PersonaRepo.PaginatePersonaCategories(req)
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminPersonaCategoryItem, 0, len(categories))
	for _, c := range categories {
		templateCount, _ := service.PersonaRepo.CountTemplatesByCategoryId(c.CategoryId)
		items = append(items, vo.AdminPersonaCategoryItem{
			CategoryId:    c.CategoryId,
			Name:          c.Name,
			Icon:          c.Icon,
			IsDefault:     c.IsDefault,
			TemplateCount: templateCount,
			Updated:       c.Updated,
		})
	}

	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

// GetAllPersonaCategories 获取所有角色分类（用于下拉选择）
func (service *PersonaService) GetAllPersonaCategories() ([]vo.AdminPersonaCategoryItem, error) {
	return service.PersonaRepo.GetAllPersonaCategories()
}

// GetPersonaCategoryDetail 角色分类详情
func (service *PersonaService) GetPersonaCategoryDetail(categoryId int) (*vo.AdminPersonaCategoryDetailRsp, error) {
	cat, err := service.PersonaRepo.GetPersonaCategoryByID(categoryId)
	if err != nil {
		return nil, errs.ErrPersonaCategoryNotFound
	}
	return &vo.AdminPersonaCategoryDetailRsp{
		CategoryId: cat.CategoryId,
		Name:       cat.Name,
		Icon:       cat.Icon,
		IsDefault:  cat.IsDefault,
		Created:    cat.Created,
		Updated:    cat.Updated,
	}, nil
}

// CreatePersonaCategory 创建角色分类
func (service *PersonaService) CreatePersonaCategory(req *vo.AdminCreatePersonaCategoryReq) error {
	cat := &po.PersonaCategory{
		Name: req.Name,
		Icon: req.Icon,
	}
	return service.PersonaRepo.CreatePersonaCategory(cat)
}

// UpdatePersonaCategory 编辑角色分类（默认分类不可编辑）
func (service *PersonaService) UpdatePersonaCategory(categoryId int, req *vo.AdminUpdatePersonaCategoryReq) error {
	cat, err := service.PersonaRepo.GetPersonaCategoryByID(categoryId)
	if err != nil {
		return errs.ErrPersonaCategoryNotFound
	}
	if cat.IsDefault == 1 {
		return errs.ErrPersonaCategoryIsDefault
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}

	if len(updates) == 0 {
		return nil
	}

	return service.PersonaRepo.UpdatePersonaCategory(categoryId, updates)
}

// DeletePersonaCategory 删除角色分类（默认分类不可删除，非默认分类删除时引用数据归属到默认分类）
func (service *PersonaService) DeletePersonaCategory(categoryId int) error {
	cat, err := service.PersonaRepo.GetPersonaCategoryByID(categoryId)
	if err != nil {
		return errs.ErrPersonaCategoryNotFound
	}
	if cat.IsDefault == 1 {
		return errs.ErrPersonaCategoryIsDefault
	}

	defaultCat, err := service.PersonaRepo.GetDefaultPersonaCategory()
	if err != nil {
		return errs.ErrPersonaCategoryNotFound
	}

	if err := service.PersonaRepo.ReassignTemplatesToCategory(categoryId, defaultCat.CategoryId); err != nil {
		return err
	}
	if err := service.PersonaRepo.ReassignUserPersonasToCategory(categoryId, defaultCat.CategoryId); err != nil {
		return err
	}

	return service.PersonaRepo.SoftDeletePersonaCategory(categoryId)
}

// batchGetCategoryMap 从模板列表中提取 categoryId 并批量查询分类
func (service *PersonaService) batchGetCategoryMap(templates []po.PersonaTemplate) map[int]po.PersonaCategory {
	ids := make([]int, 0)
	for _, t := range templates {
		if t.CategoryId > 0 {
			ids = append(ids, t.CategoryId)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	cats, err := service.PersonaRepo.BatchGetPersonaCategories(ids)
	if err != nil {
		return nil
	}

	m := make(map[int]po.PersonaCategory, len(cats))
	for _, c := range cats {
		m[c.CategoryId] = c
	}
	return m
}

// ════════════════════════════════════════════════════════════════════════════
// 用户角色
// ════════════════════════════════════════════════════════════════════════════

// validateMcpIds 校验 mcpIds 对应的 MCP 端点存在（未删除）
func (service *PersonaService) validateMcpIds(mcpIds []int) error {
	for _, id := range mcpIds {
		_, err := service.McpRepo.GetMCPEndpointByID(id)
		if err != nil {
			return errs.ErrPersonaMCPNotFound
		}
	}
	return nil
}

// validateSkillIds 校验 skillIds 对应的技能已安装（属于当前用户且 installStatus=2）
func (service *PersonaService) validateSkillIds(userId string, skillIds []int) error {
	for _, id := range skillIds {
		if _, err := service.SkillRepo.GetInstalledUserSkillByID(id, userId); err != nil {
			return errs.ErrPersonaSkillNotInstalled
		}
	}
	return nil
}

// validateAIModelId 校验 aiModelId 对应的模型存在且启用，0 表示使用默认模型跳过校验
func (service *PersonaService) validateAIModelId(aiModelId int) error {
	if aiModelId == 0 {
		return nil
	}
	model, err := service.ModelRepo.GetModelByID(aiModelId)
	if err != nil {
		return errs.ErrPersonaModelDisabled
	}
	if model.Status != 1 {
		return errs.ErrPersonaModelDisabled
	}
	return nil
}

// splitToolIds 将 persona_tool 列表按 toolType 拆分为 mcpIds 和 skillIds
func (service *PersonaService) splitToolIds(tools []po.PersonaTool) ([]int, []int) {
	mcpIds := make([]int, 0)
	skillIds := make([]int, 0)
	for _, t := range tools {
		switch t.ToolType {
		case "mcp":
			mcpIds = append(mcpIds, t.ToolId)
		case "skill":
			skillIds = append(skillIds, t.ToolId)
		}
	}
	return mcpIds, skillIds
}

// buildPersonaTools 根据 mcpIds 和 skillIds 构建 persona_tool 记录列表
func (service *PersonaService) buildPersonaTools(personaId int, userId string, mcpIds []int, skillIds []int) []po.PersonaTool {
	tools := make([]po.PersonaTool, 0, len(mcpIds)+len(skillIds))
	for _, id := range mcpIds {
		tools = append(tools, po.PersonaTool{
			PersonaId: personaId,
			UserId:    userId,
			ToolType:  "mcp",
			ToolId:    id,
		})
	}
	for _, id := range skillIds {
		tools = append(tools, po.PersonaTool{
			PersonaId: personaId,
			UserId:    userId,
			ToolType:  "skill",
			ToolId:    id,
		})
	}
	return tools
}

// ListUserPersonasSimple 获取用户角色简要信息（聊天页选择器用，仅基础字段）
func (service *PersonaService) ListUserPersonasSimple(userId string) ([]vo.UserPersonaSimpleItem, error) {
	personas, err := service.PersonaRepo.ListUserPersonasByUserId(userId)
	if err != nil {
		return nil, err
	}
	items := make([]vo.UserPersonaSimpleItem, 0, len(personas))
	for _, p := range personas {
		items = append(items, vo.UserPersonaSimpleItem{
			PersonaId: p.PersonaId,
			Name:      p.Name,
			Icon:      p.Icon,
		})
	}
	return items, nil
}

// ListUserPersonas 获取用户角色列表（含分类、模型、MCP、技能信息）
func (service *PersonaService) ListUserPersonas(userId string) ([]vo.UserPersonaListItem, error) {
	personas, err := service.PersonaRepo.ListUserPersonasByUserId(userId)
	if err != nil {
		return nil, err
	}
	if len(personas) == 0 {
		return []vo.UserPersonaListItem{}, nil
	}

	personaIds := make([]int, len(personas))
	categoryIds := make([]int, 0)
	for i, p := range personas {
		personaIds[i] = p.PersonaId
		if p.CategoryId > 0 {
			categoryIds = append(categoryIds, p.CategoryId)
		}
	}

	categoryMap := make(map[int]po.PersonaCategory)
	if len(categoryIds) > 0 {
		cats, _ := service.PersonaRepo.BatchGetPersonaCategories(categoryIds)
		for _, c := range cats {
			categoryMap[c.CategoryId] = c
		}
	}

	modelNameMap := make(map[int]string)
	for _, p := range personas {
		if p.AIModelId > 0 {
			if _, ok := modelNameMap[p.AIModelId]; !ok {
				if model, err := service.ModelRepo.GetModelByID(p.AIModelId); err == nil {
					modelNameMap[p.AIModelId] = model.ProviderDisplayName + " - " + model.DisplayName
				}
			}
		}
	}

	allTools, _ := service.PersonaRepo.BatchListPersonaToolsByPersonaIds(personaIds)
	toolMap := make(map[int][]po.PersonaTool)
	allMcpIds := make([]int, 0)
	allSkillIds := make([]int, 0)
	for _, t := range allTools {
		toolMap[t.PersonaId] = append(toolMap[t.PersonaId], t)
		if t.ToolType == "mcp" {
			allMcpIds = append(allMcpIds, t.ToolId)
		} else if t.ToolType == "skill" {
			allSkillIds = append(allSkillIds, t.ToolId)
		}
	}

	mcpNameMap := make(map[int]string)
	if len(allMcpIds) > 0 {
		mcps, _ := service.McpRepo.GetMCPEndpointsByIDs(allMcpIds)
		for _, m := range mcps {
			mcpNameMap[m.McpId] = m.DisplayName
			if mcpNameMap[m.McpId] == "" {
				mcpNameMap[m.McpId] = m.Name
			}
		}
	}

	skillNameMap := make(map[int]string)
	if len(allSkillIds) > 0 {
		skillNameMap, _ = service.SkillRepo.GetUserSkillNameMapByIDs(userId, allSkillIds)
	}

	items := make([]vo.UserPersonaListItem, 0, len(personas))
	for _, p := range personas {
		item := vo.UserPersonaListItem{
			PersonaId: p.PersonaId,
			Name:      p.Name,
			Icon:      p.Icon,
			RoleInfo:  p.RoleInfo,
			CategoryId: p.CategoryId,
			McpNames:   []string{},
			SkillNames: []string{},
			McpIds:     []int{},
			SkillIds:   []int{},
		}
		if cat, ok := categoryMap[p.CategoryId]; ok {
			item.CategoryName = cat.Name
			item.CategoryIcon = cat.Icon
		}
		if mn, ok := modelNameMap[p.AIModelId]; ok {
			item.ModelName = mn
		}
		if tools, ok := toolMap[p.PersonaId]; ok {
			for _, t := range tools {
				if t.ToolType == "mcp" {
					item.McpIds = append(item.McpIds, t.ToolId)
					if name, ok := mcpNameMap[t.ToolId]; ok {
						item.McpNames = append(item.McpNames, name)
					}
				} else if t.ToolType == "skill" {
					item.SkillIds = append(item.SkillIds, t.ToolId)
					if name, ok := skillNameMap[t.ToolId]; ok {
						item.SkillNames = append(item.SkillNames, name)
					}
				}
			}
		}
		items = append(items, item)
	}
	return items, nil
}

// GetUserPersona 获取用户角色详情（编辑页回填）
func (service *PersonaService) GetUserPersona(personaId int, userId string) (*vo.UserPersonaDetailRsp, error) {
	persona, err := service.PersonaRepo.GetUserPersonaByID(personaId, userId)
	if err != nil {
		return nil, errs.ErrPersonaNotFound
	}

	// 从 persona_tool 关联表读取 MCP/技能 ID
	tools, _ := service.PersonaRepo.ListPersonaToolsByPersona(personaId)

	// 校验并清理无效的工具关联：MCP 须存在且启用未删除，技能须已安装
	validTools := make([]po.PersonaTool, 0, len(tools))
	for _, t := range tools {
		switch t.ToolType {
		case "mcp":
			mcp, err := service.McpRepo.GetMCPEndpointByID(t.ToolId)
			if err != nil || mcp.Status != po.MCPEndpointStatusEnabled {
				// MCP 不存在、已删除或已禁用，清理 persona_tool 记录
				_ = service.PersonaRepo.DeletePersonaTool(personaId, "mcp", t.ToolId)
				continue
			}
		case "skill":
			if _, err := service.SkillRepo.GetInstalledUserSkillByID(t.ToolId, userId); err != nil {
				// 技能未安装或不存在，清理 persona_tool 记录
				_ = service.PersonaRepo.DeletePersonaTool(personaId, "skill", t.ToolId)
				continue
			}
		}
		validTools = append(validTools, t)
	}

	mcpIds, skillIds := service.splitToolIds(validTools)

	rsp := &vo.UserPersonaDetailRsp{
		PersonaId:  persona.PersonaId,
		Name:       persona.Name,
		Icon:       persona.Icon,
		RoleInfo:   persona.RoleInfo,
		CategoryId: persona.CategoryId,
		McpIds:     mcpIds,
		SkillIds:   skillIds,
		AIModelId:  persona.AIModelId,
		Created:    persona.Created,
		Updated:    persona.Updated,
	}

	// 关联分类信息
	if persona.CategoryId > 0 {
		if cat, err := service.PersonaRepo.GetPersonaCategoryByID(persona.CategoryId); err == nil {
			rsp.CategoryName = cat.Name
			rsp.CategoryIcon = cat.Icon
		}
	}

	// 关联模型名称
	if persona.AIModelId > 0 {
		if model, err := service.ModelRepo.GetModelByID(persona.AIModelId); err == nil {
			rsp.ModelName = model.ProviderDisplayName + " - " + model.DisplayName
			rsp.ModelIcon = model.Icon
		}
	}

	return rsp, nil
}

// CreateUserPersona 创建用户角色
func (service *PersonaService) CreateUserPersona(userId string, req *vo.CreateUserPersonaReq, mcpService *McpService, skillService *SkillService) error {
	// 校验分类存在
	if _, err := service.PersonaRepo.GetPersonaCategoryByID(req.CategoryId); err != nil {
		return errs.ErrPersonaCategoryNotFound
	}

	// 校验 mcpIds：每个 MCP 端点须存在且启用
	if err := service.validateMcpIds(req.McpIds); err != nil {
		return err
	}

	// 校验 skillIds：每个技能须已安装
	if err := service.validateSkillIds(userId, req.SkillIds); err != nil {
		return err
	}

	// 校验 MCP 工具名称冲突
	if err := mcpService.CheckMCPToolNameConflicts(req.McpIds); err != nil {
		return err
	}

	// 校验技能名称冲突
	if err := skillService.CheckSkillNameConflicts(userId, req.SkillIds); err != nil {
		return err
	}

	// 校验 aiModelId：模型须存在且启用（0 表示默认模型，跳过）
	if err := service.validateAIModelId(req.AIModelId); err != nil {
		return err
	}

	// 校验角色设定长度
	if utf8.RuneCountInString(req.RoleInfo) > 500 {
		return errs.ErrPersonaRoleInfoTooLong
	}

	// 如果选了模板，校验模板存在并递增使用次数
	if req.TemplateId != nil && *req.TemplateId > 0 {
		if _, err := service.PersonaRepo.GetPersonaTemplateByIDForUser(*req.TemplateId); err != nil {
			return errs.ErrPersonaTemplateNotFound
		}
		_ = service.PersonaRepo.IncrementUsageCount(*req.TemplateId)
	}

	persona := &po.UserPersona{
		UserId:     userId,
		Name:       req.Name,
		Icon:       req.Icon,
		RoleInfo:   req.RoleInfo,
		CategoryId: req.CategoryId,
		AIModelId:  req.AIModelId,
	}
	if err := service.PersonaRepo.CreateUserPersona(persona); err != nil {
		return err
	}

	// 写入 persona_tool 关联记录
	return service.PersonaRepo.BatchCreatePersonaTools(service.buildPersonaTools(persona.PersonaId, userId, req.McpIds, req.SkillIds))
}

// UpdateUserPersona 编辑用户角色
func (service *PersonaService) UpdateUserPersona(personaId int, userId string, req *vo.UpdateUserPersonaReq, mcpService *McpService, skillService *SkillService) error {
	if _, err := service.PersonaRepo.GetUserPersonaByID(personaId, userId); err != nil {
		return errs.ErrPersonaNotFound
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.RoleInfo != nil {
		if *req.RoleInfo == "" {
			return errs.ErrPersonaRoleInfoRequired
		}
		if utf8.RuneCountInString(*req.RoleInfo) > 500 {
			return errs.ErrPersonaRoleInfoTooLong
		}
		updates["role_info"] = *req.RoleInfo
	}
	if req.CategoryId != nil {
		if _, err := service.PersonaRepo.GetPersonaCategoryByID(*req.CategoryId); err != nil {
			return errs.ErrPersonaCategoryNotFound
		}
		updates["category_id"] = *req.CategoryId
	}
	if req.McpIds != nil {
		if err := service.validateMcpIds(*req.McpIds); err != nil {
			return err
		}
		// 校验 MCP 工具名称冲突
		if err := mcpService.CheckMCPToolNameConflicts(*req.McpIds); err != nil {
			return err
		}
		// 先删旧 MCP 关联，再写入新的
		_ = service.PersonaRepo.DeletePersonaToolsByPersonaAndType(personaId, "mcp")
		_ = service.PersonaRepo.BatchCreatePersonaTools(service.buildPersonaTools(personaId, userId, *req.McpIds, nil))
	}
	if req.SkillIds != nil {
		if err := service.validateSkillIds(userId, *req.SkillIds); err != nil {
			return err
		}
		// 校验技能名称冲突
		if err := skillService.CheckSkillNameConflicts(userId, *req.SkillIds); err != nil {
			return err
		}
		// 先删旧技能关联，再写入新的
		_ = service.PersonaRepo.DeletePersonaToolsByPersonaAndType(personaId, "skill")
		_ = service.PersonaRepo.BatchCreatePersonaTools(service.buildPersonaTools(personaId, userId, nil, *req.SkillIds))
	}
	if req.AIModelId != nil {
		if err := service.validateAIModelId(*req.AIModelId); err != nil {
			return err
		}
		updates["ai_model_id"] = *req.AIModelId
	}

	if len(updates) == 0 {
		return nil
	}

	return service.PersonaRepo.UpdateUserPersona(personaId, userId, updates)
}

// DeleteUserPersona 删除用户角色（软删除，关联 session.personaId 置零，清除 persona_tool 关联）
func (service *PersonaService) DeleteUserPersona(personaId int, userId string) error {
	if _, err := service.PersonaRepo.GetUserPersonaByID(personaId, userId); err != nil {
		return errs.ErrPersonaNotFound
	}

	// 将关联会话的 personaId 置零
	_ = service.PersonaRepo.ClearSessionPersonaId(personaId, userId)

	// 清除 persona_tool 关联记录
	_ = service.PersonaRepo.DeletePersonaToolsByPersona(personaId)

	return service.PersonaRepo.SoftDeleteUserPersona(personaId, userId)
}

// ListPersonaTemplatesForUser 获取用户可选角色模板列表
func (service *PersonaService) ListPersonaTemplatesForUser(categoryId *int) ([]vo.UserPersonaTemplateItem, error) {
	templates, err := service.PersonaRepo.ListPersonaTemplatesForUser(categoryId)
	if err != nil {
		return nil, err
	}

	// 批量查询分类信息
	categoryMap := service.batchGetCategoryMapFromTemplates(templates)

	items := make([]vo.UserPersonaTemplateItem, 0, len(templates))
	for _, t := range templates {
		item := vo.UserPersonaTemplateItem{
			TemplateId:  t.TemplateId,
			Name:        t.Name,
			Icon:        t.Icon,
			Description: t.Description,
			CategoryId:  t.CategoryId,
		}
		if cat, ok := categoryMap[t.CategoryId]; ok {
			item.CategoryName = cat.Name
			item.CategoryIcon = cat.Icon
		}
		items = append(items, item)
	}
	return items, nil
}

// GetPersonaTemplateDetailForUser 获取角色模板详情（含完整 roleInfo，用于预填角色设定）
func (service *PersonaService) GetPersonaTemplateDetailForUser(templateId int) (*vo.UserPersonaTemplateDetailRsp, error) {
	tmpl, err := service.PersonaRepo.GetPersonaTemplateByIDForUser(templateId)
	if err != nil {
		return nil, errs.ErrPersonaTemplateNotFound
	}

	rsp := &vo.UserPersonaTemplateDetailRsp{
		TemplateId:  tmpl.TemplateId,
		Name:        tmpl.Name,
		Icon:        tmpl.Icon,
		Description: tmpl.Description,
		RoleInfo:    tmpl.RoleInfo,
		CategoryId:  tmpl.CategoryId,
	}

	if tmpl.CategoryId > 0 {
		if cat, err := service.PersonaRepo.GetPersonaCategoryByID(tmpl.CategoryId); err == nil {
			rsp.CategoryName = cat.Name
			rsp.CategoryIcon = cat.Icon
		}
	}

	return rsp, nil
}

// ListPersonaCategoriesForUser 获取角色分类列表（用户端，按 sortOrder 排序）
func (service *PersonaService) ListPersonaCategoriesForUser() ([]vo.UserPersonaCategoryItem, error) {
	categories, err := service.PersonaRepo.ListPersonaCategoriesForUser()
	if err != nil {
		return nil, err
	}

	items := make([]vo.UserPersonaCategoryItem, 0, len(categories))
	for _, c := range categories {
		items = append(items, vo.UserPersonaCategoryItem{
			CategoryId: c.CategoryId,
			Name:       c.Name,
			Icon:       c.Icon,
		})
	}
	return items, nil
}

// batchGetCategoryMapFromTemplates 从模板列表中提取 categoryId 并批量查询分类
func (service *PersonaService) batchGetCategoryMapFromTemplates(templates []po.PersonaTemplate) map[int]po.PersonaCategory {
	return service.batchGetCategoryMap(templates)
}
