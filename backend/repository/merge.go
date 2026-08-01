package repository

import (
	"goraven/backend/po"

	"gorm.io/gorm"
)

// Merge 自动迁移数据库表结构
func Merge(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&po.Message{},
		&po.Session{},
		&po.FileLink{},
		&po.User{},
		&po.UserAuth{},
		&po.AIModel{},
		&po.AIModelMember{},
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
		&po.SkillShare{},
		&po.TeamProject{},
		&po.TeamProjectMember{},
	); err != nil {
		return err
	}

	return nil
}
