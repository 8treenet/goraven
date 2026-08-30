package repository

import (
	"goraven/backend/po"
	"goraven/backend/repository/seed"
	"goraven/config"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Seed 初始化所有种子数据
func Seed(db *gorm.DB) error {
	if err := seedSkillCategories(db); err != nil {
		return err
	}
	if err := seedPersonaCategories(db); err != nil {
		return err
	}
	if err := seedPersonaTemplates(db); err != nil {
		return err
	}

	if err := seedSkillMarket(db); err != nil {
		return err
	}
	return nil
}

func SeedOnBoot(db *gorm.DB) error {
	if err := seedSystemSettings(db); err != nil {
		return err
	}
	if err := seedSystemSkills(db); err != nil {
		return err
	}
	return nil
}

func seedSystemSettings(db *gorm.DB) error {
	keys := make([]string, len(seed.SystemSettings))
	for i, s := range seed.SystemSettings {
		keys[i] = s.Key
	}

	var existingKeys []string
	if err := db.Model(&po.SystemSetting{}).Where("config_key IN ?", keys).Pluck("config_key", &existingKeys).Error; err != nil {
		return err
	}

	existingSet := make(map[string]bool, len(existingKeys))
	for _, k := range existingKeys {
		existingSet[k] = true
	}

	var toInsert []po.SystemSetting
	for _, s := range seed.SystemSettings {
		if !existingSet[s.Key] {
			toInsert = append(toInsert, s)
		}
	}

	if len(toInsert) == 0 {
		return nil
	}
	return db.Create(&toInsert).Error
}

func seedSkillCategories(db *gorm.DB) error {
	var count int64
	if err := db.Model(&po.SkillCategory{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	lang := config.Get().GetLanguage()

	categories := make([]po.SkillCategory, 0, len(seed.SkillCategories))
	for _, s := range seed.SkillCategories {
		name := s.ZhName
		if lang == "en" {
			name = s.EnName
		}
		categories = append(categories, po.SkillCategory{Name: name, IsDefault: s.IsDefault})
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&categories).Error
}

func seedPersonaCategories(db *gorm.DB) error {
	var count int64
	if err := db.Model(&po.PersonaCategory{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	lang := config.Get().GetLanguage()

	categories := make([]po.PersonaCategory, 0, len(seed.PersonaCategories))
	for _, s := range seed.PersonaCategories {
		name := s.ZhName
		if lang == "en" {
			name = s.EnName
		}
		categories = append(categories, po.PersonaCategory{Name: name, Icon: s.Icon, IsDefault: s.IsDefault})
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&categories).Error
}

func seedPersonaTemplates(db *gorm.DB) error {
	var count int64
	if err := db.Model(&po.PersonaTemplate{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	lang := config.Get().GetLanguage()
	templates := make([]po.PersonaTemplate, 0, len(seed.PersonaTemplates))
	for _, s := range seed.PersonaTemplates {
		name := s.ZhName
		description := s.ZhDescription
		roleInfo := s.ZhRoleInfo
		categoryName := s.ZhCategoryName
		fallbackCategoryName := s.EnCategoryName
		if lang == "en" {
			name = s.EnName
			description = s.EnDescription
			roleInfo = s.EnRoleInfo
			categoryName = s.EnCategoryName
			fallbackCategoryName = s.ZhCategoryName
		}

		templates = append(templates, po.PersonaTemplate{
			Name:        name,
			Description: description,
			Icon:        s.Icon,
			RoleInfo:    roleInfo,
			CategoryId:  personaCategoryIdByName(db, categoryName, fallbackCategoryName),
			SortOrder:   s.SortOrder,
		})
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&templates).Error
}

func personaCategoryIdByName(db *gorm.DB, names ...string) int {
	var category po.PersonaCategory
	if err := db.Where("name IN ? AND deleted = 0", names).First(&category).Error; err != nil {
		return 0
	}
	return category.CategoryId
}

func seedSkillMarket(db *gorm.DB) error {
	var count int64
	if err := db.Model(&po.SkillMarket{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&seed.SkillMarketSeeds).Error
}

func seedSystemSkills(db *gorm.DB) error {
	installContent := seed.SystemSkillInstall
	installDescription := seed.ParseSkillDescription(installContent)
	chartContent := seed.SystemSkillChart
	chartDescription := seed.ParseSkillDescription(chartContent)
	guideContent := seed.SystemSkillGoRavenGuide
	guideDescription := seed.ParseSkillDescription(guideContent)
	runtimeContent := seed.SystemSkillGoRavenRuntime
	runtimeDescription := seed.ParseSkillDescription(runtimeContent)
	featuresContent := seed.SystemSkillGoRavenFeatures
	featuresDescription := seed.ParseSkillDescription(featuresContent)
	userUIContent := seed.SystemSkillGoRavenUserUI
	userUIDescription := seed.ParseSkillDescription(userUIContent)
	adminUIContent := seed.SystemSkillGoRavenAdminUI
	adminUIDescription := seed.ParseSkillDescription(adminUIContent)
	llmwikiContent := seed.SystemSkillGoRavenLLMWiki
	llmwikiDescription := seed.ParseSkillDescription(llmwikiContent)
	automationContent := seed.SystemSkillGoRavenAutomation
	automationDescription := seed.ParseSkillDescription(automationContent)
	if config.Get().System.Language == "en" {
		installContent = seed.SystemSkillInstallEn
		installDescription = seed.ParseSkillDescription(installContent)
		chartContent = seed.SystemSkillChartEn
		chartDescription = seed.ParseSkillDescription(chartContent)
		guideContent = seed.SystemSkillGoRavenGuideEn
		guideDescription = seed.ParseSkillDescription(guideContent)
		runtimeContent = seed.SystemSkillGoRavenRuntimeEn
		runtimeDescription = seed.ParseSkillDescription(runtimeContent)
		featuresContent = seed.SystemSkillGoRavenFeaturesEn
		featuresDescription = seed.ParseSkillDescription(featuresContent)
		userUIContent = seed.SystemSkillGoRavenUserUIEn
		userUIDescription = seed.ParseSkillDescription(userUIContent)
		adminUIContent = seed.SystemSkillGoRavenAdminUIEn
		adminUIDescription = seed.ParseSkillDescription(adminUIContent)
		llmwikiContent = seed.SystemSkillGoRavenLLMWikiEn
		llmwikiDescription = seed.ParseSkillDescription(llmwikiContent)
		automationContent = seed.SystemSkillGoRavenAutomationEn
		automationDescription = seed.ParseSkillDescription(automationContent)
	}

	skills := []po.SystemSkill{
		{
			Name:        "goraven-install-skill",
			Description: installDescription,
			Content:     installContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-chart",
			Description: chartDescription,
			Content:     chartContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-guide",
			Description: guideDescription,
			Content:     guideContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-runtime",
			Description: runtimeDescription,
			Content:     runtimeContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-features",
			Description: featuresDescription,
			Content:     featuresContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-user-ui",
			Description: userUIDescription,
			Content:     userUIContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-admin-ui",
			Description: adminUIDescription,
			Content:     adminUIContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-llmwiki",
			Description: llmwikiDescription,
			Content:     llmwikiContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
		{
			Name:        "goraven-automation",
			Description: automationDescription,
			Content:     automationContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		},
	}

	// docling 技能仅 Docker 环境创建（依赖容器内 Python venv + docling）
	if config.Get().Behavior.Docker {
		docContent := seed.SystemSkillGoRavenDocParse
		docDescription := seed.ParseSkillDescription(docContent)
		if config.Get().System.Language == "en" {
			docContent = seed.SystemSkillGoRavenDocParseEn
			docDescription = seed.ParseSkillDescription(docContent)
		}
		skills = append(skills, po.SystemSkill{
			Name:        "goraven-doc-parse",
			Description: docDescription,
			Content:     docContent,
			Version:     config.Version,
			Status:      po.SystemSkillStatusEnabled,
		})
	}

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}

	var existing []po.SystemSkill
	if err := db.Select("name", "version").Where("name IN ?", names).Find(&existing).Error; err != nil {
		return err
	}

	existingMap := make(map[string]string, len(existing))
	for _, e := range existing {
		existingMap[e.Name] = e.Version
	}

	var toInsert []po.SystemSkill
	for _, s := range skills {
		oldVersion, exists := existingMap[s.Name]
		if !exists {
			toInsert = append(toInsert, s)
			continue
		}

		if oldVersion != config.Version {
			if err := db.Model(&po.SystemSkill{}).Where("name = ?", s.Name).Updates(map[string]interface{}{
				"description": s.Description,
				"content":     s.Content,
				"version":     config.Version,
			}).Error; err != nil {
				continue
			}
		}
	}

	if len(toInsert) == 0 {
		return nil
	}
	return db.Create(&toInsert).Error
}
