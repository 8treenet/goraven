package repository

import (
	"raven/backend/po"

	"gorm.io/gorm"
)

func Merge(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&po.Message{},
		&po.Session{},
		&po.FileLink{},
		&po.User{},
		&po.UserAuth{},
		&po.AIModel{},
		&po.MCPEndpoint{},
		&po.SkillMarket{},
		&po.SkillCategory{},
		&po.SystemSkill{},
		&po.UserSkill{},
		&po.SystemSetting{},
		&po.UserDailyStats{},
		&po.UserPersona{},
		&po.PersonaTemplate{},
		&po.PersonaCategory{},
		&po.ShareLink{},
		&po.ChunkUpload{},
		&po.ToolDailyStats{},
		&po.PersonaTool{},
	); err != nil {
		return err
	}

	return nil
}
