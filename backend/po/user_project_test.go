package po

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserProjectTableNameAndTimestamps(t *testing.T) {
	project := UserProject{}

	if got := project.TableName(); got != "user_project" {
		t.Fatalf("TableName() = %q, want %q", got, "user_project")
	}

	if err := project.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate() error = %v", err)
	}
	if project.Created.IsZero() || project.Updated.IsZero() {
		t.Fatal("BeforeCreate() must set both timestamps")
	}
	if !project.Created.Equal(project.Updated) {
		t.Fatalf("BeforeCreate() timestamps differ: created=%v updated=%v", project.Created, project.Updated)
	}

	previousUpdated := time.Unix(1, 0)
	project.Updated = previousUpdated
	if err := project.BeforeSave(nil); err != nil {
		t.Fatalf("BeforeSave() error = %v", err)
	}
	if !project.Updated.After(previousUpdated) {
		t.Fatalf("BeforeSave() updated timestamp = %v, want after %v", project.Updated, previousUpdated)
	}
}

func TestUserProjectUpdatedPersistsOnGormUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&UserProject{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	project := &UserProject{
		UserId:      "user-1",
		ProjectName: "demo",
		Description: "before",
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	project.Description = "after"
	if err := db.Model(project).Update("description", project.Description).Error; err != nil {
		t.Fatalf("update: %v", err)
	}

	var persisted UserProject
	if err := db.First(&persisted, project.Id).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if !persisted.Updated.Equal(project.Updated) {
		t.Fatalf("persisted Updated = %v, in-memory Updated = %v", persisted.Updated, project.Updated)
	}
}
