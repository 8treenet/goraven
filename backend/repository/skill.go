package repository

import (
	"fmt"
	"raven/backend/po"
	"raven/backend/vo"
	"raven/core/iface"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *SkillRepository {
			return &SkillRepository{}
		})
	})
}

var _ iface.SystemSkillProvider = new(SkillRepository)

type SkillRepository struct {
	freedom.Repository
}

func (repo *SkillRepository) SystemSkillList() ([]iface.SkillInfo, error) {
	var skills []po.SystemSkill
	err := repo.db().Where("status = ? AND deleted = ?", po.SystemSkillStatusEnabled, 0).
		Order("skill_id asc").
		Find(&skills).Error
	if err != nil {
		return nil, err
	}

	result := make([]iface.SkillInfo, len(skills))
	for i, s := range skills {
		result[i] = iface.SkillInfo{
			Name:		s.Name,
			Description:	s.Description,
			Content:	s.Content,
		}
	}
	return result, nil
}

func (repo *SkillRepository) PaginateSystemSkills(req *vo.AdminSystemSkillListReq) ([]po.SystemSkill, *PageResult, error) {
	query := repo.db().Model(&po.SystemSkill{}).Where("deleted = 0")
	if req.Search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var skills []po.SystemSkill
	pr, err := Paginate(query.Order("updated DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &skills)
	if err != nil {
		return nil, nil, err
	}
	return skills, pr, nil
}

func (repo *SkillRepository) GetSystemSkillByID(skillId int) (*po.SystemSkill, error) {
	var skill po.SystemSkill
	err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error
	return &skill, err
}

func (repo *SkillRepository) FindSystemSkillByName(name string) (*po.SystemSkill, error) {
	var skill po.SystemSkill
	err := repo.db().First(&skill, "name = ? AND deleted = 0", name).Error
	return &skill, err
}

func (repo *SkillRepository) CreateSystemSkill(skill *po.SystemSkill) error {
	return repo.db().Create(skill).Error
}

func (repo *SkillRepository) UpdateSystemSkill(skillId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.SystemSkill{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(updates).Error
}

func (repo *SkillRepository) SoftDeleteSystemSkill(skillId int) error {
	var skill po.SystemSkill
	if err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", skill.Name, time.Now().Unix())
	return repo.db().Model(&po.SystemSkill{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(map[string]interface{}{
		"deleted":	1,
		"name":		suffixedName,
		"updated":	time.Now(),
	}).Error
}

func (repo *SkillRepository) PaginateMarketSkills(req *vo.AdminMarketSkillListReq) ([]po.SkillMarket, *PageResult, error) {
	query := repo.db().Model(&po.SkillMarket{}).Where("deleted = 0")
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.Source != "" {
		query = query.Where("source = ?", req.Source)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var skills []po.SkillMarket
	pr, err := Paginate(query.Order("sort_order ASC, updated DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &skills)
	if err != nil {
		return nil, nil, err
	}
	return skills, pr, nil
}

func (repo *SkillRepository) GetMarketSkillByID(skillId int) (*po.SkillMarket, error) {
	var skill po.SkillMarket
	err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error
	return &skill, err
}

func (repo *SkillRepository) FindMarketSkillByName(name string) (*po.SkillMarket, error) {
	var skill po.SkillMarket
	err := repo.db().First(&skill, "name = ? AND deleted = 0", name).Error
	return &skill, err
}

func (repo *SkillRepository) CreateMarketSkill(skill *po.SkillMarket) error {
	return repo.db().Create(skill).Error
}

func (repo *SkillRepository) UpdateMarketSkill(skillId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.SkillMarket{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(updates).Error
}

func (repo *SkillRepository) IncrMarketSkillInstalledCount(skillId int) error {
	return repo.db().Model(&po.SkillMarket{}).Where("skill_id = ? AND deleted = 0", skillId).
		UpdateColumn("installed_count", gorm.Expr("installed_count + ?", 1)).Error
}

func (repo *SkillRepository) SoftDeleteMarketSkill(skillId int) error {
	var skill po.SkillMarket
	if err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", skill.Name, time.Now().Unix())
	return repo.db().Model(&po.SkillMarket{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(map[string]interface{}{
		"deleted":	1,
		"name":		suffixedName,
		"updated":	time.Now(),
	}).Error
}

func (repo *SkillRepository) PaginateMarketSkillUsers(skillId int, query *PageQuery) ([]po.UserSkill, *PageResult, error) {
	db := repo.db().Model(&po.UserSkill{}).Where("market_skill_id = ?", skillId)

	var userSkills []po.UserSkill
	pr, err := Paginate(db.Order("created DESC"), query, &userSkills)
	if err != nil {
		return nil, nil, err
	}
	return userSkills, pr, nil
}

func (repo *SkillRepository) DeleteUserSkillsBySkillId(skillId int) error {
	return repo.db().Where("market_skill_id = ?", skillId).Delete(&po.UserSkill{}).Error
}

func (repo *SkillRepository) ListUserSkillIdsByMarketSkillId(marketSkillId int) ([]int, error) {
	var ids []int
	err := repo.db().Model(&po.UserSkill{}).Where("market_skill_id = ?", marketSkillId).Pluck("user_skill_id", &ids).Error
	return ids, err
}

func (repo *SkillRepository) FindInstalledUserSkills(userId string) ([]po.UserSkill, error) {
	var skills []po.UserSkill
	err := repo.db().Where("user_id = ? AND install_status = ?", userId, po.UserSkillInstalled).
		Order("created DESC").Find(&skills).Error
	return skills, err
}

func (repo *SkillRepository) FindInstalledUserSkillsByIDs(userId string, userSkillIds []int) ([]po.UserSkill, error) {
	var skills []po.UserSkill
	err := repo.db().Where("user_id = ? AND user_skill_id IN ?", userId, userSkillIds).
		Order("created DESC").Find(&skills).Error
	return skills, err
}

func (repo *SkillRepository) GetInstalledUserSkillByID(userSkillId int, userId string) (*po.UserSkill, error) {
	var skill po.UserSkill
	err := repo.db().First(&skill, "user_skill_id = ? AND user_id = ? AND install_status = ?", userSkillId, userId, po.UserSkillInstalled).Error
	return &skill, err
}

func (repo *SkillRepository) PaginateSkillCategories(req *vo.AdminSkillCategoryListReq) ([]po.SkillCategory, *PageResult, error) {
	query := repo.db().Model(&po.SkillCategory{}).Where("deleted = 0")
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}

	var categories []po.SkillCategory
	pr, err := Paginate(query.Order("category_id DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &categories)
	if err != nil {
		return nil, nil, err
	}
	return categories, pr, nil
}

func (repo *SkillRepository) GetSkillCategoryByID(categoryId int) (*po.SkillCategory, error) {
	var cat po.SkillCategory
	err := repo.db().First(&cat, "category_id = ? AND deleted = 0", categoryId).Error
	return &cat, err
}

func (repo *SkillRepository) CreateSkillCategory(cat *po.SkillCategory) error {
	return repo.db().Create(cat).Error
}

func (repo *SkillRepository) UpdateSkillCategory(categoryId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.SkillCategory{}).Where("category_id = ? AND deleted = 0", categoryId).Updates(updates).Error
}

func (repo *SkillRepository) SoftDeleteSkillCategory(categoryId int) error {
	var cat po.SkillCategory
	if err := repo.db().First(&cat, "category_id = ? AND deleted = 0", categoryId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", cat.Name, time.Now().Unix())
	return repo.db().Model(&po.SkillCategory{}).Where("category_id = ? AND deleted = 0", categoryId).Updates(map[string]interface{}{
		"deleted":	1,
		"name":		suffixedName,
		"updated":	time.Now(),
	}).Error
}

func (repo *SkillRepository) CountSkillsByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.SkillMarket{}).Where("category_id = ? AND deleted = 0", categoryId).Count(&count).Error
	return int(count), err
}

func (repo *SkillRepository) CountUserSkillsByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.UserSkill{}).Where("category_id = ?", categoryId).Count(&count).Error
	return int(count), err
}

func (repo *SkillRepository) GetDefaultSkillCategory() (*po.SkillCategory, error) {
	var cat po.SkillCategory
	err := repo.db().First(&cat, "is_default = 1 AND deleted = 0").Error
	return &cat, err
}

func (repo *SkillRepository) ReassignSkillsToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.SkillMarket{}).Where("category_id = ? AND deleted = 0", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId, "updated": time.Now()}).Error
}

func (repo *SkillRepository) ReassignUserSkillsToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.UserSkill{}).Where("category_id = ?", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId}).Error
}

func (repo *SkillRepository) GetAllSkillCategories() ([]vo.AdminSkillCategoryItem, error) {
	var categories []po.SkillCategory
	err := repo.db().Where("deleted = 0").Order("category_id DESC").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminSkillCategoryItem, 0, len(categories))
	for _, c := range categories {
		items = append(items, vo.AdminSkillCategoryItem{
			CategoryId:	c.CategoryId,
			Name:		c.Name,
			Icon:		c.Icon,
			IsDefault:	c.IsDefault,
		})
	}
	return items, nil
}

func (repo *SkillRepository) BatchGetSkillCategories(ids []int) ([]po.SkillCategory, error) {
	var cats []po.SkillCategory
	err := repo.db().Where("category_id IN ?", ids).Find(&cats).Error
	return cats, err
}

func (repo *SkillRepository) PaginateUserSkills(userId string, req *vo.UserSkillListReq) ([]po.UserSkill, *PageResult, error) {
	query := repo.db().Model(&po.UserSkill{}).Where("user_id = ?", userId)
	if req.Search != "" {
		query = query.Where("skill_name LIKE ?", "%"+req.Search+"%")
	}
	if req.CategoryId != nil {
		query = query.Where("category_id = ?", *req.CategoryId)
	}
	if req.Source != "" {
		query = query.Where("source = ?", req.Source)
	}
	if req.Status != nil {
		query = query.Where("install_status = ?", *req.Status)
	}

	var skills []po.UserSkill
	pr, err := Paginate(query.Order("updated DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &skills)
	if err != nil {
		return nil, nil, err
	}
	return skills, pr, nil
}

func (repo *SkillRepository) GetUserSkillByID(userSkillId int, userId string) (*po.UserSkill, error) {
	var skill po.UserSkill
	err := repo.db().First(&skill, "user_skill_id = ? AND user_id = ?", userSkillId, userId).Error
	return &skill, err
}

func (repo *SkillRepository) CreateUserSkill(skill *po.UserSkill) error {
	return repo.db().Create(skill).Error
}

func (repo *SkillRepository) UpdateUserSkill(userSkillId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.UserSkill{}).Where("user_skill_id = ?", userSkillId).Updates(updates).Error
}

func (repo *SkillRepository) FindUserSkillByUserIdAndName(userId, skillName string) (*po.UserSkill, error) {
	var skill po.UserSkill
	err := repo.db().First(&skill, "user_id = ? AND skill_name = ?", userId, skillName).Error
	return &skill, err
}

func (repo *SkillRepository) DeleteUserSkill(userSkillId int, userId string) error {
	return repo.db().Where("user_skill_id = ? AND user_id = ?", userSkillId, userId).Delete(&po.UserSkill{}).Error
}

func (repo *SkillRepository) PaginateUserMarketSkills(req *vo.UserMarketSkillListReq) ([]po.SkillMarket, *PageResult, error) {
	query := repo.db().Model(&po.SkillMarket{}).Where("status = ? AND deleted = ?", po.SkillStatusEnabled, 0)
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.CategoryId != nil {
		query = query.Where("category_id = ?", *req.CategoryId)
	}
	if req.Source != "" {
		query = query.Where("source = ?", req.Source)
	}

	var skills []po.SkillMarket
	pr, err := Paginate(query.Order("sort_order ASC, updated DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &skills)
	if err != nil {
		return nil, nil, err
	}
	return skills, pr, nil
}

func (repo *SkillRepository) FindAllUserSkillNames(userId string) ([]string, error) {
	var names []string
	err := repo.db().Model(&po.UserSkill{}).Where("user_id = ?", userId).Pluck("skill_name", &names).Error
	return names, err
}

func (repo *SkillRepository) GetAllCategories() ([]po.SkillCategory, error) {
	var categories []po.SkillCategory
	err := repo.db().Where("deleted = 0").Order("category_id ASC").Find(&categories).Error
	return categories, err
}

func (repo *SkillRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

func (repo *SkillRepository) GetUserSkillNamesByIDs(userId string, skillIds []int) ([]string, error) {
	if len(skillIds) == 0 {
		return nil, nil
	}
	var names []string
	err := repo.db().Model(&po.UserSkill{}).
		Where("user_skill_id IN ? AND user_id = ? AND install_status = ?", skillIds, userId, po.UserSkillInstalled).
		Pluck("skill_name", &names).Error
	return names, err
}

func (repo *SkillRepository) FindAlwaysOnSkillNames(userId string) ([]string, error) {
	var names []string
	err := repo.db().Model(&po.UserSkill{}).
		Where("user_id = ? AND install_status = ? AND always_on = ?", userId, po.UserSkillInstalled, 1).
		Pluck("skill_name", &names).Error
	return names, err
}

func (repo *SkillRepository) GetUserSkillNameMapByIDs(userId string, skillIds []int) (map[int]string, error) {
	if len(skillIds) == 0 {
		return map[int]string{}, nil
	}
	var results []struct {
		UserSkillId	int
		SkillName	string
	}
	err := repo.db().Model(&po.UserSkill{}).
		Select("user_skill_id, skill_name").
		Where("user_skill_id IN ? AND user_id = ? AND install_status = ?", skillIds, userId, po.UserSkillInstalled).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, len(results))
	for _, r := range results {
		m[r.UserSkillId] = r.SkillName
	}
	return m, nil
}
