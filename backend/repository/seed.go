package repository

import (
	"raven/backend/po"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	if err := seedSystemSettings(db); err != nil {
		return err
	}
	if err := seedSkillCategories(db); err != nil {
		return err
	}
	if err := seedPersonaCategories(db); err != nil {
		return err
	}
	if err := seedPersonaTemplates(db); err != nil {
		return err
	}
	if err := seedSystemSkills(db); err != nil {
		return err
	}
	if err := seedSkillMarket(db); err != nil {
		return err
	}
	return nil
}

func seedSystemSettings(db *gorm.DB) error {
	return nil
}

func seedSkillCategories(db *gorm.DB) error {

	return nil
}

func seedPersonaCategories(db *gorm.DB) error {
	var count int64
	if err := db.Model(&po.PersonaCategory{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return nil
}

func seedPersonaTemplates(db *gorm.DB) error {
	return nil
}

func personaCategoryIdByName(db *gorm.DB, names ...string) int {
	var category po.PersonaCategory
	if err := db.Where("name IN ? AND deleted = 0", names).First(&category).Error; err != nil {
		return 0
	}
	return category.CategoryId
}

func seedSkillMarket(db *gorm.DB) error {
	return nil
}

func seedSystemSkills(db *gorm.DB) error {
	return nil
}
