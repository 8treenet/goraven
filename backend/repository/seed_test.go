package repository

import (
	"strings"
	"testing"

	"goraven/backend/po"
	"goraven/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedSystemSkillsUsesEnglishWhenConfigured(t *testing.T) {
	cfg := config.Get()
	originalLanguage := cfg.System.Language
	cfg.System.Language = "en"
	t.Cleanup(func() {
		cfg.System.Language = originalLanguage
	})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	if err := db.AutoMigrate(&po.SystemSkill{}); err != nil {
		t.Fatalf("migrate system skill table: %v", err)
	}

	if err := seedSystemSkills(db); err != nil {
		t.Fatalf("seed system skills: %v", err)
	}

	var installSkill po.SystemSkill
	if err := db.Where("name = ?", "goraven-install-skill").First(&installSkill).Error; err != nil {
		t.Fatalf("load install skill: %v", err)
	}
	if installSkill.Description != "Guide users through installing skills in chat: create SKILL.md, configure directory structure and environment variables" {
		t.Fatalf("unexpected install skill description: %q", installSkill.Description)
	}
	if !strings.Contains(installSkill.Content, "# Skill Installation") {
		t.Fatalf("expected English install skill content, got: %q", installSkill.Content)
	}

	var chartSkill po.SystemSkill
	if err := db.Where("name = ?", "goraven-chart").First(&chartSkill).Error; err != nil {
		t.Fatalf("load chart skill: %v", err)
	}
	if chartSkill.Description != "Generate statistical charts with <goraven-chart> tags" {
		t.Fatalf("unexpected chart skill description: %q", chartSkill.Description)
	}
	if !strings.Contains(chartSkill.Content, "# Statistical Chart Generation") {
		t.Fatalf("expected English chart skill content, got: %q", chartSkill.Content)
	}
}

func TestSeedPersonaTemplatesUsesEnglishWhenConfigured(t *testing.T) {
	cfg := config.Get()
	originalLanguage := cfg.System.Language
	cfg.System.Language = "en"
	t.Cleanup(func() {
		cfg.System.Language = originalLanguage
	})

	db := newPersonaTemplateSeedTestDB(t)
	if err := seedPersonaCategories(db); err != nil {
		t.Fatalf("seed persona categories: %v", err)
	}
	if err := seedPersonaTemplates(db); err != nil {
		t.Fatalf("seed persona templates: %v", err)
	}

	var templates []po.PersonaTemplate
	if err := db.Order("sortOrder ASC").Find(&templates).Error; err != nil {
		t.Fatalf("load persona templates: %v", err)
	}
	if len(templates) != 4 {
		t.Fatalf("expected 4 persona templates, got %d", len(templates))
	}

	expectedNames := []string{"General Assistant", "Data Analyst", "Financial Research Analyst", "Operations Strategy Advisor"}
	for i, name := range expectedNames {
		if templates[i].Name != name {
			t.Fatalf("template %d name = %q, want %q", i, templates[i].Name, name)
		}
		if templates[i].Description == "" {
			t.Fatalf("template %q should have an English description", name)
		}
		if !strings.Contains(templates[i].RoleInfo, "You are") {
			t.Fatalf("template %q should have English role info, got %q", name, templates[i].RoleInfo)
		}
	}

	var dataCategory po.PersonaCategory
	if err := db.Where("name = ?", "Data Analysis").First(&dataCategory).Error; err != nil {
		t.Fatalf("load data analysis category: %v", err)
	}
	if templates[1].CategoryId != dataCategory.CategoryId {
		t.Fatalf("data analyst categoryId = %d, want %d", templates[1].CategoryId, dataCategory.CategoryId)
	}
}

func TestSeedPersonaTemplatesUsesChineseByDefault(t *testing.T) {
	cfg := config.Get()
	originalLanguage := cfg.System.Language
	cfg.System.Language = "zh"
	t.Cleanup(func() {
		cfg.System.Language = originalLanguage
	})

	db := newPersonaTemplateSeedTestDB(t)
	if err := seedPersonaCategories(db); err != nil {
		t.Fatalf("seed persona categories: %v", err)
	}
	if err := seedPersonaTemplates(db); err != nil {
		t.Fatalf("seed persona templates: %v", err)
	}

	var templates []po.PersonaTemplate
	if err := db.Order("sortOrder ASC").Find(&templates).Error; err != nil {
		t.Fatalf("load persona templates: %v", err)
	}
	if len(templates) != 4 {
		t.Fatalf("expected 4 persona templates, got %d", len(templates))
	}

	expectedNames := []string{"通用助手", "数据分析师", "金融研究分析师", "运营策略顾问"}
	for i, name := range expectedNames {
		if templates[i].Name != name {
			t.Fatalf("template %d name = %q, want %q", i, templates[i].Name, name)
		}
		if templates[i].Description == "" {
			t.Fatalf("template %q should have a Chinese description", name)
		}
		if !strings.Contains(templates[i].RoleInfo, "你是") {
			t.Fatalf("template %q should have Chinese role info, got %q", name, templates[i].RoleInfo)
		}
	}

	var businessCategory po.PersonaCategory
	if err := db.Where("name = ?", "商业效率").First(&businessCategory).Error; err != nil {
		t.Fatalf("load business efficiency category: %v", err)
	}
	if templates[2].CategoryId != businessCategory.CategoryId || templates[3].CategoryId != businessCategory.CategoryId {
		t.Fatalf("financial and operations templates should use business category %d, got %d and %d", businessCategory.CategoryId, templates[2].CategoryId, templates[3].CategoryId)
	}
}

func newPersonaTemplateSeedTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	if err := db.AutoMigrate(&po.PersonaCategory{}, &po.PersonaTemplate{}); err != nil {
		t.Fatalf("migrate persona seed tables: %v", err)
	}
	return db
}
