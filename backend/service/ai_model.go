package service

import (
	"context"
	"encoding/json"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/repository/seed"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/core/provider"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/schema"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *AIModelService {
			return &AIModelService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *AIModelService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type AIModelService struct {
	Worker        freedom.Worker
	ModelRepo     *repository.ProviderRepository
	DashboardRepo *repository.DashboardRepository
}

func (svc *AIModelService) ListProviders() []vo.ProviderItem {
	lang := config.Get().GetLanguage()
	items := make([]vo.ProviderItem, 0, len(seed.ProviderDefs))
	for _, def := range seed.ProviderDefs {
		var defaultBaseURL string
		if lang == "en" {
			defaultBaseURL = def.DefaultBaseURLEn
		} else {
			defaultBaseURL = def.DefaultBaseURLZh
		}
		items = append(items, vo.ProviderItem{
			ProviderID:            def.ID,
			ProviderDisplayNameZh: def.DisplayNameZh,
			ProviderDisplayNameEn: def.DisplayNameEn,
			Icon:                  def.Icon,
			DefaultBaseURL:        defaultBaseURL,
			RequireAPIKey:         def.RequireAPIKey,
			RequireBaseURL:        def.RequireBaseURL,
		})
	}
	return items
}

func (svc *AIModelService) ListModels(req *vo.AdminModelListReq) (*infra.PageResponse, error) {
	items, pr, err := svc.ModelRepo.PaginateModels(req)
	if err != nil {
		return nil, err
	}
	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

func (service *AIModelService) ListEnabledModels() ([]vo.UserModelItem, error) {
	models, err := service.ModelRepo.FindEnabledModels()
	if err != nil {
		return nil, err
	}

	items := make([]vo.UserModelItem, 0, len(models))
	for _, m := range models {
		items = append(items, vo.UserModelItem{
			AIModelId:           m.AIModelId,
			ProviderDisplayName: m.ProviderDisplayName,
			DisplayName:         m.DisplayName,
			ModelName:           m.ModelName,
			Icon:                m.Icon,
			ContextLen:          m.ContextLen,
			IsDefault:           m.IsDefault,
			IsFlash:             m.IsFlash,
			IsVisual:            m.IsVisual,
		})
	}
	return items, nil
}

func (service *AIModelService) GetEnabledModelByID(id int) (*vo.UserModelItem, error) {
	model, err := service.ModelRepo.GetModelByID(id)
	if err != nil {
		return nil, nil
	}

	return &vo.UserModelItem{
		AIModelId:           model.AIModelId,
		ProviderDisplayName: model.ProviderDisplayName,
		DisplayName:         model.DisplayName,
		ModelName:           model.ModelName,
		Icon:                model.Icon,
		ContextLen:          model.ContextLen,
		IsDefault:           model.IsDefault,
		IsFlash:             model.IsFlash,
		IsVisual:            model.IsVisual,
	}, nil
}

func (svc *AIModelService) CreateModel(req *vo.AdminCreateModelReq) error {
	def := svc.findProviderDef(req.ProviderID)
	if def == nil {
		return errs.ErrProviderNotFound
	}

	if def.RequireAPIKey && req.APIKey == "" {
		return errs.ErrAPIKeyRequired
	}
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = def.CurrentDefaultBaseURL()
	}
	if def.RequireBaseURL && baseURL == "" {
		return errs.ErrBaseURLRequired
	}
	if err := svc.validateExtraFields(req.ExtraFields); err != nil {
		return err
	}

	if err := svc.testModelKeys(req.ProviderID, req.APIKey, baseURL, req.ExtraFields, req.ProxyURL, req.ModelName, req.ContextLen); err != nil {
		return err
	}

	contextLen := req.ContextLen
	if contextLen <= 0 {
		contextLen = 200
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.ModelName
	}

	model := &po.AIModel{
		ProviderDisplayName: req.ProviderDisplayName,
		DisplayName:         displayName,
		ProviderID:          req.ProviderID,
		ModelName:           req.ModelName,
		Icon:                req.Icon,
		APIKey:              req.APIKey,
		BaseURL:             baseURL,
		ExtraFields:         req.ExtraFields,
		ProxyURL:            req.ProxyURL,
		ContextLen:          contextLen,
		IsDefault:           req.IsDefault,
		IsFlash:             req.IsFlash,
		IsVisual:            req.IsVisual,
		Status:              1,
		Remark:              req.Remark,
	}

	if err := svc.ModelRepo.CreateModel(model); err != nil {
		return err
	}
	svc.invalidateDashboardCache()
	return nil
}

func (svc *AIModelService) UpdateModel(id int, req *vo.AdminUpdateModelReq) error {
	existing, err := svc.ModelRepo.GetModelByID(id)
	if err != nil {
		return errs.ErrModelNotFound
	}

	updates := map[string]interface{}{}
	needTest := false

	if req.ProviderDisplayName != "" {
		updates["provider_display_name"] = req.ProviderDisplayName
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	} else if req.ModelName != "" {
		updates["display_name"] = req.ModelName
	}
	if req.ModelName != "" {
		updates["model_name"] = req.ModelName
		if req.ModelName != existing.ModelName {
			needTest = true
		}
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
		if req.APIKey != existing.APIKey {
			needTest = true
		}
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
		if req.BaseURL != existing.BaseURL {
			needTest = true
		}
	}
	if req.ExtraFields != "" {
		if err := svc.validateExtraFields(req.ExtraFields); err != nil {
			return err
		}
		updates["extra_fields"] = req.ExtraFields
	}
	if req.ProxyURL != "" {
		updates["proxy_url"] = req.ProxyURL
		if req.ProxyURL != existing.ProxyURL {
			needTest = true
		}
	}
	if req.ContextLen > 0 {
		updates["context_len"] = req.ContextLen
	}
	if req.IsDefault != nil {
		updates["is_default"] = int(*req.IsDefault)
	}
	if req.IsFlash != nil {
		updates["is_flash"] = *req.IsFlash
	}
	if req.IsVisual != nil {
		updates["is_visual"] = *req.IsVisual
	}

	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	if len(updates) == 0 {
		return nil
	}

	if needTest {
		apiKey := existing.APIKey
		if req.APIKey != "" {
			apiKey = req.APIKey
		}
		baseURL := existing.BaseURL
		if req.BaseURL != "" {
			baseURL = req.BaseURL
		}
		proxyURL := existing.ProxyURL
		if req.ProxyURL != "" {
			proxyURL = req.ProxyURL
		}
		modelName := existing.ModelName
		if req.ModelName != "" {
			modelName = req.ModelName
		}
		contextLen := existing.ContextLen
		if req.ContextLen > 0 {
			contextLen = req.ContextLen
		}
		if err := svc.testModelKeys(existing.ProviderID, apiKey, baseURL, existing.ExtraFields, proxyURL, modelName, contextLen); err != nil {
			return err
		}
	}

	if err := svc.ModelRepo.UpdateModel(id, updates); err != nil {
		return err
	}
	svc.invalidateDashboardCache()
	return nil
}

func (svc *AIModelService) DeleteModel(id int) error {
	existing, err := svc.ModelRepo.GetModelByID(id)
	if err != nil {
		return errs.ErrModelNotFound
	}
	if existing.IsDefault == 1 {
		return errs.ErrCannotDeleteDefaultModel
	}
	if existing.IsFlash == 1 {
		return errs.ErrCannotDeleteFlashModel
	}
	if existing.IsVisual == 1 {
		return errs.ErrCannotDeleteVisualModel
	}
	if err := svc.ModelRepo.SoftDeleteModel(id); err != nil {
		return err
	}
	if svc.DashboardRepo != nil {
		svc.DashboardRepo.InvalidateAllDashboardCache()
	}
	return nil
}

func (svc *AIModelService) UpdateModelStatus(id int, status uint8) error {
	if _, err := svc.ModelRepo.GetModelByID(id); err != nil {
		return errs.ErrModelNotFound
	}
	if err := svc.ModelRepo.UpdateModel(id, map[string]interface{}{"status": status}); err != nil {
		return err
	}
	svc.invalidateDashboardCache()
	return nil
}

func (svc *AIModelService) GetModelDetail(id int) (*vo.AdminModelDetailRsp, error) {
	model, err := svc.ModelRepo.GetModelByID(id)
	if err != nil {
		return nil, errs.ErrModelNotFound
	}
	return &vo.AdminModelDetailRsp{
		AIModelId:           model.AIModelId,
		ProviderDisplayName: model.ProviderDisplayName,
		DisplayName:         model.DisplayName,
		ProviderID:          model.ProviderID,
		ModelName:           model.ModelName,
		Icon:                model.Icon,
		APIKey:              model.APIKey,
		BaseURL:             model.BaseURL,
		ExtraFields:         model.ExtraFields,
		ProxyURL:            model.ProxyURL,
		ContextLen:          model.ContextLen,
		IsDefault:           model.IsDefault,
		IsFlash:             model.IsFlash,
		IsVisual:            model.IsVisual,
		Status:              model.Status,
		Remark:              model.Remark,
		Created:             model.Created,
		Updated:             model.Updated,
	}, nil
}

func (svc *AIModelService) SetDefaultModel(id int) error {
	if _, err := svc.ModelRepo.GetModelByID(id); err != nil {
		return errs.ErrModelNotFound
	}
	if err := svc.ModelRepo.UpdateModel(id, map[string]interface{}{"is_default": 1}); err != nil {
		return err
	}
	svc.invalidateDashboardCache()
	return nil
}

func (svc *AIModelService) SetFlashModel(id int) error {
	if _, err := svc.ModelRepo.GetModelByID(id); err != nil {
		return errs.ErrModelNotFound
	}
	if err := svc.ModelRepo.UpdateModel(id, map[string]interface{}{"is_flash": 1}); err != nil {
		return err
	}
	svc.invalidateDashboardCache()
	return nil
}

func (svc *AIModelService) SetVisualModel(id int) error {
	if _, err := svc.ModelRepo.GetModelByID(id); err != nil {
		return errs.ErrModelNotFound
	}
	if err := svc.ModelRepo.UpdateModel(id, map[string]interface{}{"is_visual": 1}); err != nil {
		return err
	}
	svc.invalidateDashboardCache()
	return nil
}

func (svc *AIModelService) invalidateDashboardCache() {
	if svc.DashboardRepo != nil {
		svc.DashboardRepo.InvalidateAllDashboardCache()
	}
}

func (svc *AIModelService) RecommendModels(providerID string, apiKey string, baseURL string) ([]vo.RecommendModelItem, error) {
	def := svc.findProviderDef(providerID)
	if def == nil {
		return nil, nil
	}

	pv, err := provider.GetProviderByName(providerID, provider.ProviderConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		return nil, nil
	}

	models, err := pv.Models()
	if err != nil || models == nil {
		return nil, nil
	}

	items := make([]vo.RecommendModelItem, 0, len(models))
	for _, m := range models {
		items = append(items, vo.RecommendModelItem{
			ID:      m.ID,
			Object:  m.Object,
			OwnedBy: m.OwnedBy,
		})
	}
	return items, nil
}

func (svc *AIModelService) testModelKeys(providerID, apiKey, baseURL, extraFields, proxyURL, modelName string, contextLen int) error {
	if apiKey == "" {
		return errs.ErrAPIKeyRequired
	}

	cfg := provider.ProviderConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		ExtraFields: extraFields,
	}
	pv, err := provider.GetProviderByName(providerID, cfg)
	if err != nil {
		return errs.NewFormatError("provider error: %v", "供应商错误: %v", err)
	}
	if proxyURL != "" {
		pv.SetProxy(proxyURL)
	}
	return svc.testSingleKey(pv, modelName, contextLen)
}

func (svc *AIModelService) testSingleKey(pv provider.Provider, modelName string, contextLen int) error {
	chatModel, err := pv.CreateModel(modelName, false, contextLen*1024)
	if err != nil {
		return errs.NewFormatError("create model error: %v", "创建模型错误: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := chatModel.Generate(ctx, []*schema.Message{
		{Role: schema.User, Content: "hi"},
	})
	if err != nil {
		return errs.NewFormatError("model test failed: %v", "模型测试失败: %v", err)
	}

	if result == nil || result.Content == "" {
		return errs.ErrModelTestFailed
	}

	return nil
}

func (svc *AIModelService) findProviderDef(id string) *seed.ProviderDef {
	for i := range seed.ProviderDefs {
		if seed.ProviderDefs[i].ID == id {
			return &seed.ProviderDefs[i]
		}
	}
	return nil
}

func (svc *AIModelService) validateExtraFields(raw string) error {
	if raw == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return errs.NewFormatError(
			"extraFields invalid JSON: %v, example: {\"thinking\":{\"type\":\"enabled\"}}",
			"extraFields JSON格式错误: %v, 示例: {\"thinking\":{\"type\":\"enabled\"}}",
			err,
		)
	}
	return nil
}
