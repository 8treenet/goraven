package repository

import (
	"fmt"
	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/core/iface"
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

// SkillRepository 技能仓库
type SkillRepository struct {
	freedom.Repository
}

// SystemSkillList 获取启用的系统技能列表（供 agent 加载使用）
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
			Name:        s.Name,
			Description: s.Description,
			Content:     s.Content,
		}
	}
	return result, nil
}

// PaginateSystemSkills 管理员系统技能分页列表
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

// GetSystemSkillByID 根据 ID 查询系统技能详情
func (repo *SkillRepository) GetSystemSkillByID(skillId int) (*po.SystemSkill, error) {
	var skill po.SystemSkill
	err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error
	return &skill, err
}

// FindSystemSkillByName 根据名称查询未删除的系统技能
func (repo *SkillRepository) FindSystemSkillByName(name string) (*po.SystemSkill, error) {
	var skill po.SystemSkill
	err := repo.db().First(&skill, "name = ? AND deleted = 0", name).Error
	return &skill, err
}

// CreateSystemSkill 创建系统技能
func (repo *SkillRepository) CreateSystemSkill(skill *po.SystemSkill) error {
	return repo.db().Create(skill).Error
}

// UpdateSystemSkill 更新系统技能
func (repo *SkillRepository) UpdateSystemSkill(skillId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.SystemSkill{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(updates).Error
}

// SoftDeleteSystemSkill 软删除系统技能，name 追加时间戳后缀以允许复用
func (repo *SkillRepository) SoftDeleteSystemSkill(skillId int) error {
	var skill po.SystemSkill
	if err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", skill.Name, time.Now().Unix())
	return repo.db().Model(&po.SystemSkill{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(map[string]interface{}{
		"deleted": 1,
		"name":    suffixedName,
		"updated": time.Now(),
	}).Error
}

// PaginateMarketSkills 市场技能分页列表（仅查 skill_market 表）
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

// GetMarketSkillByID 根据 ID 查询未删除的市场技能
func (repo *SkillRepository) GetMarketSkillByID(skillId int) (*po.SkillMarket, error) {
	var skill po.SkillMarket
	err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error
	return &skill, err
}

// FindMarketSkillByName 根据名称查询未删除的市场技能（用于唯一性校验）
func (repo *SkillRepository) FindMarketSkillByName(name string) (*po.SkillMarket, error) {
	var skill po.SkillMarket
	err := repo.db().First(&skill, "name = ? AND deleted = 0", name).Error
	return &skill, err
}

// CreateMarketSkill 创建市场技能记录
func (repo *SkillRepository) CreateMarketSkill(skill *po.SkillMarket) error {
	return repo.db().Create(skill).Error
}

// UpdateMarketSkill 部分更新市场技能
func (repo *SkillRepository) UpdateMarketSkill(skillId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.SkillMarket{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(updates).Error
}

// IncrMarketSkillInstalledCount 市场技能安装计数 +1
func (repo *SkillRepository) IncrMarketSkillInstalledCount(skillId int) error {
	return repo.db().Model(&po.SkillMarket{}).Where("skill_id = ? AND deleted = 0", skillId).
		UpdateColumn("installed_count", gorm.Expr("installed_count + ?", 1)).Error
}

// SoftDeleteMarketSkill 软删除市场技能，name 追加时间戳后缀以释放唯一索引
func (repo *SkillRepository) SoftDeleteMarketSkill(skillId int) error {
	var skill po.SkillMarket
	if err := repo.db().First(&skill, "skill_id = ? AND deleted = 0", skillId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", skill.Name, time.Now().Unix())
	return repo.db().Model(&po.SkillMarket{}).Where("skill_id = ? AND deleted = 0", skillId).Updates(map[string]interface{}{
		"deleted": 1,
		"name":    suffixedName,
		"updated": time.Now(),
	}).Error
}

// PaginateMarketSkillUsers 查询安装了该市场技能的用户分页列表
func (repo *SkillRepository) PaginateMarketSkillUsers(skillId int, query *PageQuery) ([]po.UserSkill, *PageResult, error) {
	db := repo.db().Model(&po.UserSkill{}).Where("market_skill_id = ?", skillId)

	var userSkills []po.UserSkill
	pr, err := Paginate(db.Order("created DESC"), query, &userSkills)
	if err != nil {
		return nil, nil, err
	}
	return userSkills, pr, nil
}

// DeleteUserSkillsBySkillId 删除指定技能的所有用户安装记录
func (repo *SkillRepository) DeleteUserSkillsBySkillId(skillId int) error {
	return repo.db().Where("market_skill_id = ?", skillId).Delete(&po.UserSkill{}).Error
}

// ListUserSkillIdsByMarketSkillId 获取指定市场技能的所有用户安装记录ID
func (repo *SkillRepository) ListUserSkillIdsByMarketSkillId(marketSkillId int) ([]int, error) {
	var ids []int
	err := repo.db().Model(&po.UserSkill{}).Where("market_skill_id = ?", marketSkillId).Pluck("user_skill_id", &ids).Error
	return ids, err
}

// FindInstalledUserSkills 查询指定用户已安装成功的技能列表
func (repo *SkillRepository) FindInstalledUserSkills(userId string) ([]po.UserSkill, error) {
	var skills []po.UserSkill
	err := repo.db().Where("user_id = ? AND install_status = ?", userId, po.UserSkillInstalled).
		Order("created DESC").Find(&skills).Error
	return skills, err
}

// FindInstalledUserSkillsByIDs 查询指定用户已安装成功的技能列表（按 userSkillId 筛选）
func (repo *SkillRepository) FindInstalledUserSkillsByIDs(userId string, userSkillIds []int) ([]po.UserSkill, error) {
	var skills []po.UserSkill
	err := repo.db().Where("user_id = ? AND user_skill_id IN ?", userId, userSkillIds).
		Order("created DESC").Find(&skills).Error
	return skills, err
}

// GetInstalledUserSkillByID 根据 ID 查询用户已安装的技能
func (repo *SkillRepository) GetInstalledUserSkillByID(userSkillId int, userId string) (*po.UserSkill, error) {
	var skill po.UserSkill
	err := repo.db().First(&skill, "user_skill_id = ? AND user_id = ? AND install_status = ?", userSkillId, userId, po.UserSkillInstalled).Error
	return &skill, err
}

// ════════════════════════════════════════════════════════════════════════════
// 技能分类相关方法
// ════════════════════════════════════════════════════════════════════════════

// PaginateSkillCategories 技能分类分页列表
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

// GetSkillCategoryByID 根据 ID 查询分类
func (repo *SkillRepository) GetSkillCategoryByID(categoryId int) (*po.SkillCategory, error) {
	var cat po.SkillCategory
	err := repo.db().First(&cat, "category_id = ? AND deleted = 0", categoryId).Error
	return &cat, err
}

// CreateSkillCategory 创建分类
func (repo *SkillRepository) CreateSkillCategory(cat *po.SkillCategory) error {
	return repo.db().Create(cat).Error
}

// UpdateSkillCategory 更新分类
func (repo *SkillRepository) UpdateSkillCategory(categoryId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.SkillCategory{}).Where("category_id = ? AND deleted = 0", categoryId).Updates(updates).Error
}

// SoftDeleteSkillCategory 软删除分类（name 追加时间戳后缀）
func (repo *SkillRepository) SoftDeleteSkillCategory(categoryId int) error {
	var cat po.SkillCategory
	if err := repo.db().First(&cat, "category_id = ? AND deleted = 0", categoryId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", cat.Name, time.Now().Unix())
	return repo.db().Model(&po.SkillCategory{}).Where("category_id = ? AND deleted = 0", categoryId).Updates(map[string]interface{}{
		"deleted": 1,
		"name":    suffixedName,
		"updated": time.Now(),
	}).Error
}

// CountSkillsByCategoryId 统计分类下技能数量
func (repo *SkillRepository) CountSkillsByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.SkillMarket{}).Where("category_id = ? AND deleted = 0", categoryId).Count(&count).Error
	return int(count), err
}

// CountUserSkillsByCategoryId 统计分类下用户安装记录数量
func (repo *SkillRepository) CountUserSkillsByCategoryId(categoryId int) (int, error) {
	var count int64
	err := repo.db().Model(&po.UserSkill{}).Where("category_id = ?", categoryId).Count(&count).Error
	return int(count), err
}

// GetDefaultSkillCategory 获取默认技能分类
func (repo *SkillRepository) GetDefaultSkillCategory() (*po.SkillCategory, error) {
	var cat po.SkillCategory
	err := repo.db().First(&cat, "is_default = 1 AND deleted = 0").Error
	return &cat, err
}

// ReassignSkillsToCategory 将指定分类下的市场技能归属到目标分类
func (repo *SkillRepository) ReassignSkillsToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.SkillMarket{}).Where("category_id = ? AND deleted = 0", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId, "updated": time.Now()}).Error
}

// ReassignUserSkillsToCategory 将指定分类下的用户安装记录归属到目标分类
func (repo *SkillRepository) ReassignUserSkillsToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.UserSkill{}).Where("category_id = ?", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId}).Error
}

// ReassignSkillSharesToCategory 将指定分类下的共享技能记录归属到目标分类
func (repo *SkillRepository) ReassignSkillSharesToCategory(oldCategoryId, newCategoryId int) error {
	return repo.db().Model(&po.SkillShare{}).Where("category_id = ?", oldCategoryId).
		Updates(map[string]interface{}{"category_id": newCategoryId}).Error
}

// GetAllSkillCategories 获取所有分类（用于下拉选择）
func (repo *SkillRepository) GetAllSkillCategories() ([]vo.AdminSkillCategoryItem, error) {
	var categories []po.SkillCategory
	err := repo.db().Where("deleted = 0").Order("category_id DESC").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminSkillCategoryItem, 0, len(categories))
	for _, c := range categories {
		items = append(items, vo.AdminSkillCategoryItem{
			CategoryId: c.CategoryId,
			Name:       c.Name,
			Icon:       c.Icon,
			IsDefault:  c.IsDefault,
		})
	}
	return items, nil
}

// BatchGetSkillCategories 根据 ID 列表批量查询分类（供 service 层组装使用）
func (repo *SkillRepository) BatchGetSkillCategories(ids []int) ([]po.SkillCategory, error) {
	var cats []po.SkillCategory
	err := repo.db().Where("category_id IN ?", ids).Find(&cats).Error
	return cats, err
}

// ════════════════════════════════════════════════════════════════════════════
// 用户技能相关方法
// ════════════════════════════════════════════════════════════════════════════

// PaginateUserSkills 分页查询用户技能列表（含搜索/分类/来源/状态筛选）
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

// GetUserSkillByID 根据ID查询用户技能（含userId校验）
func (repo *SkillRepository) GetUserSkillByID(userSkillId int, userId string) (*po.UserSkill, error) {
	var skill po.UserSkill
	err := repo.db().First(&skill, "user_skill_id = ? AND user_id = ?", userSkillId, userId).Error
	return &skill, err
}

// CreateUserSkill 创建用户技能记录
func (repo *SkillRepository) CreateUserSkill(skill *po.UserSkill) error {
	return repo.db().Create(skill).Error
}

// UpdateUserSkill 更新用户技能指定字段
func (repo *SkillRepository) UpdateUserSkill(userSkillId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.UserSkill{}).Where("user_skill_id = ?", userSkillId).Updates(updates).Error
}

// FindUserSkillByUserIdAndName 根据userId+skillName查询用户技能（安装时判重）
func (repo *SkillRepository) FindUserSkillByUserIdAndName(userId, skillName string) (*po.UserSkill, error) {
	var skill po.UserSkill
	err := repo.db().First(&skill, "user_id = ? AND skill_name = ?", userId, skillName).Error
	return &skill, err
}

// DeleteUserSkill 删除用户技能记录
func (repo *SkillRepository) DeleteUserSkill(userSkillId int, userId string) error {
	return repo.db().Where("user_skill_id = ? AND user_id = ?", userSkillId, userId).Delete(&po.UserSkill{}).Error
}

// PaginateUserMarketSkills 分页查询市场技能（用户视角，仅上架且未删除）
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

// FindAllUserSkillNames 查询用户所有已安装技能的 skillName（用于市场列表判断 userInstalled）
func (repo *SkillRepository) FindAllUserSkillNames(userId string) ([]string, error) {
	var names []string
	err := repo.db().Model(&po.UserSkill{}).Where("user_id = ?", userId).Pluck("skill_name", &names).Error
	return names, err
}

// GetAllCategories 获取所有技能分类（未删除）
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

// GetUserSkillNamesByIDs 根据技能ID列表批量获取已安装的用户技能名称
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

// FindAlwaysOnSkillNames 获取用户始终启用的已安装技能名称
func (repo *SkillRepository) FindAlwaysOnSkillNames(userId string) ([]string, error) {
	var names []string
	err := repo.db().Model(&po.UserSkill{}).
		Where("user_id = ? AND install_status = ? AND always_on = ?", userId, po.UserSkillInstalled, 1).
		Pluck("skill_name", &names).Error
	return names, err
}

// GetUserSkillNameMapByIDs 批量获取用户技能 ID -> 名称映射
func (repo *SkillRepository) GetUserSkillNameMapByIDs(userId string, skillIds []int) (map[int]string, error) {
	if len(skillIds) == 0 {
		return map[int]string{}, nil
	}
	var results []struct {
		UserSkillId int
		SkillName   string
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

// PaginateSkillShares 团队共享技能分页列表
func (repo *SkillRepository) PaginateSkillShares(req *vo.SkillShareListReq) ([]po.SkillShare, *PageResult, error) {
	query := repo.db().Model(&po.SkillShare{})
	if req.Search != "" {
		query = query.Where("skill_name LIKE ?", "%"+req.Search+"%")
	}
	var shares []po.SkillShare
	pr, err := Paginate(query.Order("created DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &shares)
	if err != nil {
		return nil, nil, err
	}
	return shares, pr, nil
}

// CreateSkillShare 创建团队共享技能记录
func (repo *SkillRepository) CreateSkillShare(share *po.SkillShare) error {
	return repo.db().Create(share).Error
}

// GetSkillShareByID 通过 ID 查询团队共享技能
func (repo *SkillRepository) GetSkillShareByID(shareId int) (*po.SkillShare, error) {
	var share po.SkillShare
	err := repo.db().First(&share, "share_id = ?", shareId).Error
	return &share, err
}

// GetSkillShareBySkillName 通过技能名查询团队共享技能
func (repo *SkillRepository) GetSkillShareBySkillName(skillName string) (*po.SkillShare, error) {
	var share po.SkillShare
	err := repo.db().First(&share, "skill_name = ?", skillName).Error
	return &share, err
}

// DeleteSkillShare 删除团队共享技能记录
func (repo *SkillRepository) DeleteSkillShare(shareId int) error {
	return repo.db().Delete(&po.SkillShare{}, "share_id = ?", shareId).Error
}

// UpdateSkillShare 更新团队共享技能记录
func (repo *SkillRepository) UpdateSkillShare(shareId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.SkillShare{}).Where("share_id = ?", shareId).Updates(updates).Error
}

// IncrSkillShareInstallCount 递增团队共享技能的安装次数
func (repo *SkillRepository) IncrSkillShareInstallCount(shareId int) error {
	return repo.db().Model(&po.SkillShare{}).Where("share_id = ?", shareId).
		UpdateColumn("install_count", gorm.Expr("install_count + 1")).Error
}
