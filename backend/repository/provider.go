package repository

import (
	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/core/iface"
	"goraven/core/provider"
	"goraven/util"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *ProviderRepository {
			return &ProviderRepository{}
		})
	})
}

type ProviderRepository struct {
	freedom.Repository
}

func (repo *ProviderRepository) PaginateModels(req *vo.AdminModelListReq) ([]vo.AdminModelItem, *PageResult, error) {
	query := repo.db().Model(&po.AIModel{}).Where("deleted = 0")
	if req.ProviderID != "" {
		query = query.Where("provider_id = ?", req.ProviderID)
	}
	if req.Search != "" {
		query = query.Where("model_name LIKE ?", "%"+req.Search+"%")
	}

	var models []po.AIModel
	pr, err := Paginate(query.Order("created DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &models)
	if err != nil {
		return nil, nil, err
	}

	items := make([]vo.AdminModelItem, 0, len(models))
	for _, m := range models {
		items = append(items, vo.AdminModelItem{
			AIModelId:           m.AIModelId,
			ProviderDisplayName: m.ProviderDisplayName,
			DisplayName:         m.DisplayName,
			ProviderID:          m.ProviderID,
			ModelName:           m.ModelName,
			Icon:                m.Icon,
			APIKeyMasked:        maskAPIKey(m.APIKey),
			BaseURL:             m.BaseURL,
			ProxyURL:            m.ProxyURL,
			ContextLen:          m.ContextLen,
			ExtraFields:         m.ExtraFields,
			IsDefault:           m.IsDefault,
		IsFlash:             m.IsFlash,
		IsVisual:            m.IsVisual,
			Status:              m.Status,
			Access:              m.Access,
			Remark:              m.Remark,
			Created:             m.Created,
			Updated:             m.Updated,
		})
	}

	return items, pr, nil
}

func (repo *ProviderRepository) CreateModel(model *po.AIModel) error {
	return repo.db().Transaction(func(tx *gorm.DB) error {
		if model.IsDefault == 1 {
			if err := tx.Model(&po.AIModel{}).Where("is_default = 1 AND deleted = 0").Update("is_default", 0).Error; err != nil {
				return err
			}
		}
		if model.IsFlash == 1 {
			if err := tx.Model(&po.AIModel{}).Where("is_flash = 1 AND deleted = 0").Update("is_flash", 0).Error; err != nil {
				return err
			}
		}
		if model.IsVisual == 1 {
			if err := tx.Model(&po.AIModel{}).Where("is_visual = 1 AND deleted = 0").Update("is_visual", 0).Error; err != nil {
				return err
			}
		}

		var defaultCount int64
		if err := tx.Model(&po.AIModel{}).Where("is_default = 1 AND deleted = 0").Count(&defaultCount).Error; err != nil {
			return err
		}
		if defaultCount == 0 {
			model.IsDefault = 1
		}

		return tx.Create(model).Error
	})
}

func (repo *ProviderRepository) GetModelByID(id int) (*po.AIModel, error) {
	var model po.AIModel
	err := repo.db().First(&model, "ai_model_id = ? AND deleted = 0", id).Error
	return &model, err
}

func (repo *ProviderRepository) UpdateModel(id int, updates map[string]interface{}) error {
	return repo.db().Transaction(func(tx *gorm.DB) error {
		if v, ok := updates["is_default"]; ok && util.IntFromIFace(v) == 1 {
			if err := tx.Model(&po.AIModel{}).Where("is_default = 1 AND deleted = 0").Update("is_default", 0).Error; err != nil {
				return err
			}
		}
		if v, ok := updates["is_flash"]; ok && util.IntFromIFace(v) == 1 {
			if err := tx.Model(&po.AIModel{}).Where("is_flash = 1 AND deleted = 0").Update("is_flash", 0).Error; err != nil {
				return err
			}
		}
		if v, ok := updates["is_visual"]; ok && util.IntFromIFace(v) == 1 {
			if err := tx.Model(&po.AIModel{}).Where("is_visual = 1 AND deleted = 0").Update("is_visual", 0).Error; err != nil {
				return err
			}
		}

		updates["updated"] = time.Now()
		return tx.Model(&po.AIModel{}).Where("ai_model_id = ? AND deleted = 0", id).Updates(updates).Error
	})
}

// SoftDeleteModel 软删除模型，同时清空 isDefault/isFlash/isVisual 标志位
// 防止删除后仍占用「默认/Flash/视觉」角色，导致 SetDefault/SetFlash/SetVisual 链路与
// 已删除行冲突，以及仪表盘模型使用分布出现重复的「Unknown」聚合。
func (repo *ProviderRepository) SoftDeleteModel(id int) error {
	return repo.db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.AIModel{}).
			Where("ai_model_id = ? AND deleted = 0", id).
			Updates(map[string]interface{}{
				"is_default": 0,
				"is_flash":   0,
				"is_visual":  0,
				"deleted":    1,
				"updated":    time.Now(),
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

// FindEnabledModels 查询所有启用且未删除的模型
func (repo *ProviderRepository) FindEnabledModels() ([]po.AIModel, error) {
	var models []po.AIModel
	err := repo.db().Where("status = 1 AND deleted = 0").Order("created DESC").Find(&models).Error
	return models, err
}

// FindEnabledModelsByUser 查询用户可见的启用模型（全员开放 OR 成员），按创建时间倒序
func (repo *ProviderRepository) FindEnabledModelsByUser(userId string) ([]po.AIModel, error) {
	var models []po.AIModel
	err := repo.db().
		Where("status = 1 AND deleted = 0 AND (access = ? OR ai_model_id IN (?))",
			po.AIModelAccessAll,
			repo.db().Model(&po.AIModelMember{}).Select("ai_model_id").Where("user_id = ?", userId),
		).
		Order("created DESC").
		Find(&models).Error
	return models, err
}

// GetDefaultModel 获取默认模型（isDefault=1，启用，未删除）
func (repo *ProviderRepository) GetDefaultModel() (*po.AIModel, error) {
	var model po.AIModel
	err := repo.db().Where("is_default = 1 AND status = 1 AND deleted = 0").First(&model).Error
	return &model, err
}

// GetDefaultChatModel 获取默认聊天模型
// 封装 GetDefaultModel → 创建 provider → 创建 ChatModel 的完整流程
func (repo *ProviderRepository) GetDefaultChatModel() (iface.BaseChatModel, error) {
	model, err := repo.GetDefaultModel()
	if err != nil {
		return nil, err
	}
	return repo.createChatModelFromPO(model, false)
}

// GetFlashChatModel 获取 Flash 模型
// 优先级：isFlash=1 的模型 > fallbackModelId 指定的模型 > 默认模型
func (repo *ProviderRepository) GetFlashChatModel(fallbackModelId int) (iface.BaseChatModel, error) {
	// 1. 优先使用 Flash 模型
	var model po.AIModel
	if err := repo.db().Where("is_flash = 1 AND status = 1 AND deleted = 0").First(&model).Error; err == nil {
		return repo.createChatModelFromPO(&model, false)
	}

	// 2. 其次使用 fallbackModelId
	if fallbackModelId > 0 {
		if m, err := repo.GetModelByID(fallbackModelId); err == nil {
			return repo.createChatModelFromPO(m, false)
		}
	}

	// 3. 最后降级为默认模型
	return repo.GetDefaultChatModel()
}

// HasVisualModel 检查是否存在多模态识别模型（isVisual=1 且启用）
func (repo *ProviderRepository) HasVisualModel() (bool, error) {
	var count int64
	if err := repo.db().Model(&po.AIModel{}).Where("is_visual = 1 AND status = 1 AND deleted = 0").Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetVisualChatModel 获取多模态识别模型
// 仅查询 isVisual=1 的模型，不降级，找不到返回 nil
func (repo *ProviderRepository) GetVisualChatModel() (iface.BaseChatModel, error) {
	var model po.AIModel
	if err := repo.db().Where("is_visual = 1 AND status = 1 AND deleted = 0").First(&model).Error; err != nil {
		return nil, nil
	}
	return repo.createChatModelFromPO(&model, false)
}

// createChatModelFromPO 根据 po.AIModel 创建聊天模型
func (repo *ProviderRepository) createChatModelFromPO(model *po.AIModel, reasoning bool) (iface.BaseChatModel, error) {
	pv, err := provider.GetProviderByName(model.ProviderID, provider.ProviderConfig{
		APIKey:      model.APIKey,
		BaseURL:     model.BaseURL,
		ExtraFields: model.ExtraFields,
	})
	if err != nil {
		return nil, err
	}
	if model.ProxyURL != "" {
		pv.SetProxy(model.ProxyURL)
	}

	return pv.CreateModel(model.ModelName, reasoning, model.ContextLenInTokens())
}

// CreateChatModelFromID 根据模型ID创建聊天模型，支持控制推理模式
// 返回 chatModel（LLM 客户端）、aiModel（模型元数据，含 DisplayName/Icon/ContextLen）、
// aiModelId（实际使用的模型ID）和 error
func (repo *ProviderRepository) CreateChatModelFromID(aiModelId int, reasoning bool) (iface.BaseChatModel, *po.AIModel, int, error) {
	var am *po.AIModel
	var defaultModel *po.AIModel
	var defaultModelErr error

	defaultModel, defaultModelErr = repo.GetDefaultModel()
	if defaultModelErr == nil {
		am = defaultModel
	}
	useModel, useModelErr := repo.GetModelByID(aiModelId)
	if useModelErr == nil {
		am = useModel
	}
	if am == nil {
		//返回中英错误类型
		return nil, nil, 0, errs.ErrModelAndDefaultNotFound
	}

	resultModel, err := repo.createChatModelFromPO(am, reasoning)
	return resultModel, am, am.AIModelId, err
}

func (repo *ProviderRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// ListModelMembers 查询模型的所有可见成员
func (repo *ProviderRepository) ListModelMembers(modelId int) ([]po.AIModelMember, error) {
	var members []po.AIModelMember
	err := repo.db().Where("ai_model_id = ?", modelId).Order("created ASC").Find(&members).Error
	return members, err
}

// AddModelMember 添加模型可见成员
func (repo *ProviderRepository) AddModelMember(modelId int, userId string) error {
	member := &po.AIModelMember{
		AIModelId: modelId,
		UserId:    userId,
	}
	return repo.db().Create(member).Error
}

// RemoveModelMember 移除模型可见成员
func (repo *ProviderRepository) RemoveModelMember(modelId int, userId string) error {
	return repo.db().Where("ai_model_id = ? AND user_id = ?", modelId, userId).
		Delete(&po.AIModelMember{}).Error
}

// RemoveModelMembersByModelId 删除模型的所有成员记录（模型删除时调用）
func (repo *ProviderRepository) RemoveModelMembersByModelId(modelId int) error {
	return repo.db().Where("ai_model_id = ?", modelId).Delete(&po.AIModelMember{}).Error
}

// UpdateModelAccess 更新模型访问权限
func (repo *ProviderRepository) UpdateModelAccess(modelId int, access uint8) error {
	return repo.db().Model(&po.AIModel{}).Where("ai_model_id = ? AND deleted = 0", modelId).
		Update("access", access).Error
}
