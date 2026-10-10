package service

import (
	"testing"
	"time"

	"goraven/backend/po"

	"gorm.io/gorm"
)

func TestGitCloneDepthLegacyMigration(t *testing.T) {
	svc := newGitServiceTestEnv(t)
	var db *gorm.DB
	if err := svc.GitSettingRepo.FetchDB(&db); err != nil {
		t.Fatal(err)
	}
	// Recreate the old table shape with a pre-existing failed clone row.
	if err := db.Migrator().DropColumn(&po.ProjectGitSetting{}, "full_history"); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO project_git_setting
		(owner_type, owner_id, remote_url, clone_state, created, updated)
		VALUES (?, ?, ?, ?, ?, ?)`, po.GitOwnerUserProject, 111,
		"https://example.com/repo.git", po.GitCloneFailed, time.Now(), time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&po.ProjectGitSetting{}); err != nil {
		t.Fatal(err)
	}
	setting, err := svc.GitSettingRepo.GetByOwner(po.GitOwnerUserProject, 111)
	if err != nil {
		t.Fatal(err)
	}
	if setting.FullHistory {
		t.Fatal("existing rows must keep shallow-clone retry behavior")
	}
	if setting.CloneState != po.GitCloneFailed {
		t.Fatal("migration changed the existing clone state")
	}
}
