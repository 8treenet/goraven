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

type PersonaRepository struct {
	freedom.Repository
}

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

func (repo *PersonaRepository) GetPersonaTemplateByID(templateId int) (*po.PersonaTemplate, error) {
	var tmpl po.PersonaTemplate
	err := repo.db().First(&tmpl, "template_id = ? AND deleted = 0", templateId).Error
	return &tmpl, err
}

func (repo *PersonaRepository) CreatePersonaTemplate(tmpl *po.PersonaTemplate) error {
	return repo.db().Create(tmpl).Error
}

func (repo *PersonaRepository) UpdatePersonaTemplate(templateId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.PersonaTemplate{}).Where("template_id = ? AND deleted = 0", templateId).Updates(updates).Error
}

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

func (repo *PersonaRepository) IncrementUsageCount(templateId int) error {
	return repo.db().Model(&po.PersonaTemplate{}).Where("template_id = ? AND deleted = 0", templateId).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

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

func (repo *PersonaRepository) GetPersonaCategoryByID(categoryId int) (*po.PersonaCategory, error) {
	var cat po.PersonaCategory
	err := repo.db().First(&cat, "category_id = ? AND deleted = 0", categoryId).Error
	return &cat, err
}

func (repo *PersonaRepository) CreatePersonaCategory(cat *po.PersonaCategory) error {
	return repo.db().Create(cat).Error
}

func (repo *PersonaRepository) UpdatePersonaCategory(categoryId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.PersonaCategory{}).Where("category_id = ? AND deleted = 0", categoryId).Updates(updates).Error
}

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

func (repo *PersonaRepository) CountTemplatesByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.PersonaTemplate{}).Where("category_id = ? AND deleted = 0", categoryId).Count(&count).Error
	return int(count), err
}

func (repo *PersonaRepository) CountUserPersonasByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.UserPersona{}).Where("category_id = ? AND deleted = 0", categoryId).Count(&count).Error
	return int(count), err
}

func (repo *PersonaRepository) GetDefaultPersonaCategory() (*po.PersonaCategory, error) {
	var cat po.PersonaCategory
	err := repo.db().First(&cat, "is_default = 1 AND deleted = 0").Error
	return &cat, err
}

func (repo *PersonaRepository) ReassignTemplatesToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.PersonaTemplate{}).Where("category_id = ? AND deleted = 0", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId, "updated": time.Now()}).Error
}

func (repo *PersonaRepository) ReassignUserPersonasToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.UserPersona{}).Where("category_id = ? AND deleted = 0", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId, "updated": time.Now()}).Error
}

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

func (repo *PersonaRepository) BatchGetPersonaCategories(ids []int) ([]po.PersonaCategory, error) {
	var cats []po.PersonaCategory
	err := repo.db().Where("category_id IN ?", ids).Find(&cats).Error
	return cats, err
}

func (repo *PersonaRepository) ListUserPersonasByUserId(userId string) ([]po.UserPersona, error) {
	var personas []po.UserPersona
	err := repo.db().Where("user_id = ? AND deleted = 0", userId).
		Order("updated DESC").Find(&personas).Error
	return personas, err
}

func (repo *PersonaRepository) GetUserPersonaByID(personaId int, userId string) (*po.UserPersona, error) {
	var persona po.UserPersona
	err := repo.db().First(&persona, "persona_id = ? AND user_id = ? AND deleted = 0", personaId, userId).Error
	return &persona, err
}

func (repo *PersonaRepository) CreateUserPersona(persona *po.UserPersona) error {
	return repo.db().Create(persona).Error
}

func (repo *PersonaRepository) UpdateUserPersona(personaId int, userId string, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.UserPersona{}).Where("persona_id = ? AND user_id = ? AND deleted = 0", personaId, userId).Updates(updates).Error
}

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

func (repo *PersonaRepository) ClearSessionPersonaId(personaId int, userId string) error {
	return repo.db().Model(&po.Session{}).Where("persona_id = ? AND user_id = ?", personaId, userId).
		Updates(map[string]interface{}{"persona_id": 0, "updated": time.Now()}).Error
}

func (repo *PersonaRepository) ListPersonaTemplatesForUser(categoryId *int) ([]po.PersonaTemplate, error) {
	query := repo.db().Where("deleted = 0")
	if categoryId != nil {
		query = query.Where("category_id = ?", *categoryId)
	}
	var templates []po.PersonaTemplate
	err := query.Order("sort_order ASC, updated DESC").Find(&templates).Error
	return templates, err
}

func (repo *PersonaRepository) GetPersonaTemplateByIDForUser(templateId int) (*po.PersonaTemplate, error) {
	var tmpl po.PersonaTemplate
	err := repo.db().First(&tmpl, "template_id = ? AND deleted = 0", templateId).Error
	return &tmpl, err
}

func (repo *PersonaRepository) ListPersonaCategoriesForUser() ([]po.PersonaCategory, error) {
	var categories []po.PersonaCategory
	err := repo.db().Where("deleted = 0").Order("category_id ASC").Find(&categories).Error
	return categories, err
}

func (repo *PersonaRepository) BatchCreatePersonaTools(tools []po.PersonaTool) error {
	if len(tools) == 0 {
		return nil
	}
	return repo.db().Create(&tools).Error
}

func (repo *PersonaRepository) DeletePersonaToolsByPersona(personaId int) error {
	return repo.db().Where("persona_id = ?", personaId).Delete(&po.PersonaTool{}).Error
}

func (repo *PersonaRepository) DeletePersonaTool(personaId int, toolType string, toolId int) error {
	return repo.db().Where("persona_id = ? AND tool_type = ? AND tool_id = ?", personaId, toolType, toolId).Delete(&po.PersonaTool{}).Error
}

func (repo *PersonaRepository) DeletePersonaToolsByPersonaAndType(personaId int, toolType string) error {
	return repo.db().Where("persona_id = ? AND tool_type = ?", personaId, toolType).Delete(&po.PersonaTool{}).Error
}

func (repo *PersonaRepository) ListPersonaToolsByPersona(personaId int) ([]po.PersonaTool, error) {
	var tools []po.PersonaTool
	err := repo.db().Where("persona_id = ?", personaId).Find(&tools).Error
	return tools, err
}

func (repo *PersonaRepository) BatchListPersonaToolsByPersonaIds(personaIds []int) ([]po.PersonaTool, error) {
	if len(personaIds) == 0 {
		return nil, nil
	}
	var tools []po.PersonaTool
	err := repo.db().Where("persona_id IN ?", personaIds).Find(&tools).Error
	return tools, err
}

func (repo *PersonaRepository) DeletePersonaToolByMcpId(mcpId int) error {
	return repo.db().Where("tool_type = ? AND tool_id = ?", "mcp", mcpId).Delete(&po.PersonaTool{}).Error
}

func (repo *PersonaRepository) DeletePersonaToolBySkillId(userSkillId int) error {
	return repo.db().Where("tool_type = ? AND tool_id = ?", "skill", userSkillId).Delete(&po.PersonaTool{}).Error
}

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
