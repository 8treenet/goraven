package repository

import (
	"fmt"
	"goraven/backend/po"
	"goraven/backend/vo"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *PersonaRepository {
			return &PersonaRepository{}
		})
	})
}

// PersonaRepository 角色模板与分类仓库
type PersonaRepository struct {
	freedom.Repository
}

// ════════════════════════════════════════════════════════════════════════════
// 角色模板
// ════════════════════════════════════════════════════════════════════════════

// PaginatePersonaTemplates 角色模板分页列表
func (repo *PersonaRepository) PaginatePersonaTemplates(req *vo.AdminPersonaTemplateListReq) ([]po.PersonaTemplate, *PageResult, error) {
	query := repo.db().Model(&po.PersonaTemplate{}).Where("deleted = 0")
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.CategoryId != nil {
		query = query.Where("category_id = ?", *req.CategoryId)
	}

	var templates []po.PersonaTemplate
	pr, err := Paginate(query.Order("sort_order ASC, updated DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &templates)
	if err != nil {
		return nil, nil, err
	}
	return templates, pr, nil
}

// GetPersonaTemplateByID 根据 ID 查询角色模板
func (repo *PersonaRepository) GetPersonaTemplateByID(templateId int) (*po.PersonaTemplate, error) {
	var tmpl po.PersonaTemplate
	err := repo.db().First(&tmpl, "template_id = ? AND deleted = 0", templateId).Error
	return &tmpl, err
}

// CreatePersonaTemplate 创建角色模板
func (repo *PersonaRepository) CreatePersonaTemplate(tmpl *po.PersonaTemplate) error {
	return repo.db().Create(tmpl).Error
}

// UpdatePersonaTemplate 部分更新角色模板
func (repo *PersonaRepository) UpdatePersonaTemplate(templateId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.PersonaTemplate{}).Where("template_id = ? AND deleted = 0", templateId).Updates(updates).Error
}

// SoftDeletePersonaTemplate 软删除角色模板
func (repo *PersonaRepository) SoftDeletePersonaTemplate(templateId int) error {
	var tmpl po.PersonaTemplate
	if err := repo.db().First(&tmpl, "template_id = ? AND deleted = 0", templateId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", tmpl.Name, time.Now().Unix())
	return repo.db().Model(&po.PersonaTemplate{}).Where("template_id = ? AND deleted = 0", templateId).Updates(map[string]interface{}{
		"deleted": 1,
		"name":    suffixedName,
		"updated": time.Now(),
	}).Error
}

// IncrementUsageCount 模板使用次数 +1
func (repo *PersonaRepository) IncrementUsageCount(templateId int) error {
	return repo.db().Model(&po.PersonaTemplate{}).Where("template_id = ? AND deleted = 0", templateId).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

// ════════════════════════════════════════════════════════════════════════════
// 角色分类
// ════════════════════════════════════════════════════════════════════════════

// PaginatePersonaCategories 角色分类分页列表
func (repo *PersonaRepository) PaginatePersonaCategories(req *vo.AdminPersonaCategoryListReq) ([]po.PersonaCategory, *PageResult, error) {
	query := repo.db().Model(&po.PersonaCategory{}).Where("deleted = 0")
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}

	var categories []po.PersonaCategory
	pr, err := Paginate(query.Order("category_id ASC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &categories)
	if err != nil {
		return nil, nil, err
	}
	return categories, pr, nil
}

// GetPersonaCategoryByID 根据 ID 查询角色分类
func (repo *PersonaRepository) GetPersonaCategoryByID(categoryId int) (*po.PersonaCategory, error) {
	var cat po.PersonaCategory
	err := repo.db().First(&cat, "category_id = ? AND deleted = 0", categoryId).Error
	return &cat, err
}

// CreatePersonaCategory 创建角色分类
func (repo *PersonaRepository) CreatePersonaCategory(cat *po.PersonaCategory) error {
	return repo.db().Create(cat).Error
}

// UpdatePersonaCategory 更新角色分类
func (repo *PersonaRepository) UpdatePersonaCategory(categoryId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.PersonaCategory{}).Where("category_id = ? AND deleted = 0", categoryId).Updates(updates).Error
}

// SoftDeletePersonaCategory 软删除角色分类（name 追加时间戳后缀）
func (repo *PersonaRepository) SoftDeletePersonaCategory(categoryId int) error {
	var cat po.PersonaCategory
	if err := repo.db().First(&cat, "category_id = ? AND deleted = 0", categoryId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", cat.Name, time.Now().Unix())
	return repo.db().Model(&po.PersonaCategory{}).Where("category_id = ? AND deleted = 0", categoryId).Updates(map[string]interface{}{
		"deleted": 1,
		"name":    suffixedName,
		"updated": time.Now(),
	}).Error
}

// CountTemplatesByCategoryId 统计分类下模板数量
func (repo *PersonaRepository) CountTemplatesByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.PersonaTemplate{}).Where("category_id = ? AND deleted = 0", categoryId).Count(&count).Error
	return int(count), err
}

// CountUserPersonasByCategoryId 统计使用该分类的用户角色数量
func (repo *PersonaRepository) CountUserPersonasByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.UserPersona{}).Where("category_id = ? AND deleted = 0", categoryId).Count(&count).Error
	return int(count), err
}

// GetDefaultPersonaCategory 获取默认角色分类
func (repo *PersonaRepository) GetDefaultPersonaCategory() (*po.PersonaCategory, error) {
	var cat po.PersonaCategory
	err := repo.db().First(&cat, "is_default = 1 AND deleted = 0").Error
	return &cat, err
}

// ReassignTemplatesToCategory 将指定分类下的角色模板归属到目标分类
func (repo *PersonaRepository) ReassignTemplatesToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.PersonaTemplate{}).Where("category_id = ? AND deleted = 0", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId, "updated": time.Now()}).Error
}

// ReassignUserPersonasToCategory 将指定分类下的用户角色归属到目标分类
func (repo *PersonaRepository) ReassignUserPersonasToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.UserPersona{}).Where("category_id = ? AND deleted = 0", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId, "updated": time.Now()}).Error
}

// GetAllPersonaCategories 获取所有角色分类（用于下拉选择）
func (repo *PersonaRepository) GetAllPersonaCategories() ([]vo.AdminPersonaCategoryItem, error) {
	var categories []po.PersonaCategory
	err := repo.db().Where("deleted = 0").Order("category_id ASC").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminPersonaCategoryItem, 0, len(categories))
	for _, c := range categories {
		items = append(items, vo.AdminPersonaCategoryItem{
			CategoryId: c.CategoryId,
			Name:       c.Name,
			Icon:       c.Icon,
			IsDefault:  c.IsDefault,
		})
	}
	return items, nil
}

// BatchGetPersonaCategories 根据 ID 列表批量查询分类
func (repo *PersonaRepository) BatchGetPersonaCategories(ids []int) ([]po.PersonaCategory, error) {
	var cats []po.PersonaCategory
	err := repo.db().Where("category_id IN ?", ids).Find(&cats).Error
	return cats, err
}

// ════════════════════════════════════════════════════════════════════════════
// 用户角色
// ════════════════════════════════════════════════════════════════════════════

// ListUserPersonasByUserId 获取用户角色列表
func (repo *PersonaRepository) ListUserPersonasByUserId(userId string) ([]po.UserPersona, error) {
	var personas []po.UserPersona
	err := repo.db().Where("user_id = ? AND deleted = 0", userId).
		Order("updated DESC").Find(&personas).Error
	return personas, err
}

// GetUserPersonaByID 根据 ID 查询用户角色
func (repo *PersonaRepository) GetUserPersonaByID(personaId int, userId string) (*po.UserPersona, error) {
	var persona po.UserPersona
	err := repo.db().First(&persona, "persona_id = ? AND user_id = ? AND deleted = 0", personaId, userId).Error
	return &persona, err
}

// GetUserPersonasByIDs 根据 ID 列表批量获取未删除的用户角色
func (repo *PersonaRepository) GetUserPersonasByIDs(ids []int) ([]po.UserPersona, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var personas []po.UserPersona
	err := repo.db().Where("persona_id IN ? AND deleted = 0", ids).Find(&personas).Error
	return personas, err
}

// CreateUserPersona 创建用户角色
func (repo *PersonaRepository) CreateUserPersona(persona *po.UserPersona) error {
	return repo.db().Create(persona).Error
}

// UpdateUserPersona 部分更新用户角色
func (repo *PersonaRepository) UpdateUserPersona(personaId int, userId string, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.UserPersona{}).Where("persona_id = ? AND user_id = ? AND deleted = 0", personaId, userId).Updates(updates).Error
}

// SoftDeleteUserPersona 软删除用户角色（name 追加时间戳后缀，释放唯一索引）
func (repo *PersonaRepository) SoftDeleteUserPersona(personaId int, userId string) error {
	var persona po.UserPersona
	if err := repo.db().First(&persona, "persona_id = ? AND user_id = ? AND deleted = 0", personaId, userId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", persona.Name, time.Now().Unix())
	return repo.db().Model(&po.UserPersona{}).Where("persona_id = ? AND user_id = ? AND deleted = 0", personaId, userId).Updates(map[string]interface{}{
		"deleted": 1,
		"name":    suffixedName,
		"updated": time.Now(),
	}).Error
}

// ClearSessionPersonaId 删除角色时，将关联会话的 personaId 置零
func (repo *PersonaRepository) ClearSessionPersonaId(personaId int, userId string) error {
	return repo.db().Model(&po.Session{}).Where("persona_id = ? AND user_id = ?", personaId, userId).
		Updates(map[string]interface{}{"persona_id": 0, "updated": time.Now()}).Error
}

// ListPersonaTemplatesForUser 获取用户可选角色模板列表（deleted=0，支持按分类筛选）
func (repo *PersonaRepository) ListPersonaTemplatesForUser(categoryId *int) ([]po.PersonaTemplate, error) {
	query := repo.db().Where("deleted = 0")
	if categoryId != nil {
		query = query.Where("category_id = ?", *categoryId)
	}
	var templates []po.PersonaTemplate
	err := query.Order("sort_order ASC, updated DESC").Find(&templates).Error
	return templates, err
}

// GetPersonaTemplateByIDForUser 根据 ID 查询角色模板（用户端，deleted=0）
func (repo *PersonaRepository) GetPersonaTemplateByIDForUser(templateId int) (*po.PersonaTemplate, error) {
	var tmpl po.PersonaTemplate
	err := repo.db().First(&tmpl, "template_id = ? AND deleted = 0", templateId).Error
	return &tmpl, err
}

// ListPersonaCategoriesForUser 获取角色分类列表（用户端，按 categoryId 排序）
func (repo *PersonaRepository) ListPersonaCategoriesForUser() ([]po.PersonaCategory, error) {
	var categories []po.PersonaCategory
	err := repo.db().Where("deleted = 0").Order("category_id ASC").Find(&categories).Error
	return categories, err
}

// ════════════════════════════════════════════════════════════════════════════
// 角色关联工具
// ════════════════════════════════════════════════════════════════════════════

// BatchCreatePersonaTools 批量创建角色关联工具
func (repo *PersonaRepository) BatchCreatePersonaTools(tools []po.PersonaTool) error {
	if len(tools) == 0 {
		return nil
	}
	return repo.db().Create(&tools).Error
}

// DeletePersonaToolsByPersona 删除指定角色的所有关联工具
func (repo *PersonaRepository) DeletePersonaToolsByPersona(personaId int) error {
	return repo.db().Where("persona_id = ?", personaId).Delete(&po.PersonaTool{}).Error
}

// DeletePersonaTool 删除指定角色的单条关联工具记录
func (repo *PersonaRepository) DeletePersonaTool(personaId int, toolType string, toolId int) error {
	return repo.db().Where("persona_id = ? AND tool_type = ? AND tool_id = ?", personaId, toolType, toolId).Delete(&po.PersonaTool{}).Error
}

// DeletePersonaToolsByPersonaAndType 删除指定角色指定类型的所有关联工具
func (repo *PersonaRepository) DeletePersonaToolsByPersonaAndType(personaId int, toolType string) error {
	return repo.db().Where("persona_id = ? AND tool_type = ?", personaId, toolType).Delete(&po.PersonaTool{}).Error
}

// ListPersonaToolsByPersona 获取指定角色的所有关联工具
func (repo *PersonaRepository) ListPersonaToolsByPersona(personaId int) ([]po.PersonaTool, error) {
	var tools []po.PersonaTool
	err := repo.db().Where("persona_id = ?", personaId).Find(&tools).Error
	return tools, err
}

// BatchListPersonaToolsByPersonaIds 批量获取多个角色的所有关联工具
func (repo *PersonaRepository) BatchListPersonaToolsByPersonaIds(personaIds []int) ([]po.PersonaTool, error) {
	if len(personaIds) == 0 {
		return nil, nil
	}
	var tools []po.PersonaTool
	err := repo.db().Where("persona_id IN ?", personaIds).Find(&tools).Error
	return tools, err
}

// DeletePersonaToolByMcpId 删除指定 MCP 端点ID 关联的所有角色工具记录
func (repo *PersonaRepository) DeletePersonaToolByMcpId(mcpId int) error {
	return repo.db().Where("tool_type = ? AND tool_id = ?", "mcp", mcpId).Delete(&po.PersonaTool{}).Error
}

// DeletePersonaToolBySkillId 删除指定用户技能ID 关联的所有角色工具记录
func (repo *PersonaRepository) DeletePersonaToolBySkillId(userSkillId int) error {
	return repo.db().Where("tool_type = ? AND tool_id = ?", "skill", userSkillId).Delete(&po.PersonaTool{}).Error
}

// DeletePersonaToolsBySkillIds 批量删除指定用户技能ID列表关联的角色工具记录
func (repo *PersonaRepository) DeletePersonaToolsBySkillIds(userSkillIds []int) error {
	if len(userSkillIds) == 0 {
		return nil
	}
	return repo.db().Where("tool_type = ? AND tool_id IN ?", "skill", userSkillIds).Delete(&po.PersonaTool{}).Error
}

func (repo *PersonaRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
