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

func (repo *ProviderRepository) FindEnabledModels() ([]po.AIModel, error) {
	var models []po.AIModel
	err := repo.db().Where("status = 1 AND deleted = 0").Order("created DESC").Find(&models).Error
	return models, err
}

func (repo *ProviderRepository) GetDefaultModel() (*po.AIModel, error) {
	var model po.AIModel
	err := repo.db().Where("is_default = 1 AND status = 1 AND deleted = 0").First(&model).Error
	return &model, err
}

func (repo *ProviderRepository) GetDefaultChatModel() (iface.BaseChatModel, error) {
	model, err := repo.GetDefaultModel()
	if err != nil {
		return nil, err
	}
	return repo.createChatModelFromPO(model, false)
}

func (repo *ProviderRepository) GetFlashChatModel(fallbackModelId int) (iface.BaseChatModel, error) {

	var model po.AIModel
	if err := repo.db().Where("is_flash = 1 AND status = 1 AND deleted = 0").First(&model).Error; err == nil {
		return repo.createChatModelFromPO(&model, false)
	}

	if fallbackModelId > 0 {
		if m, err := repo.GetModelByID(fallbackModelId); err == nil {
			return repo.createChatModelFromPO(m, false)
		}
	}

	return repo.GetDefaultChatModel()
}

func (repo *ProviderRepository) HasVisualModel() (bool, error) {
	var count int64
	if err := repo.db().Model(&po.AIModel{}).Where("is_visual = 1 AND status = 1 AND deleted = 0").Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *ProviderRepository) GetVisualChatModel() (iface.BaseChatModel, error) {
	var model po.AIModel
	if err := repo.db().Where("is_visual = 1 AND status = 1 AND deleted = 0").First(&model).Error; err != nil {
		return nil, nil
	}
	return repo.createChatModelFromPO(&model, false)
}

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
