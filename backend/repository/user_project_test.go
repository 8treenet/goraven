package repository_test

import (
	"errors"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/backend/repository"

	"github.com/8treenet/freedom"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserProjectRepositoryPropagatesProjectReferences(t *testing.T) {
	db, repo := newUserProjectRepositoryTest(t)

	project := &po.UserProject{UserId: "user-1", ProjectName: "old-name"}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}

	matchingSession := &po.Session{
		SessionId:       "session-match",
		UserId:          "user-1",
		Project:         "old-name",
		SharedProjectId: 0,
	}
	otherUserSession := &po.Session{
		SessionId:       "session-other-user",
		UserId:          "user-2",
		Project:         "old-name",
		SharedProjectId: 0,
	}
	sharedSession := &po.Session{
		SessionId:       "session-shared",
		UserId:          "user-1",
		Project:         "old-name",
		SharedProjectId: 7,
	}
	for _, session := range []*po.Session{matchingSession, otherUserSession, sharedSession} {
		if err := db.Create(session).Error; err != nil {
			t.Fatalf("create session %q: %v", session.SessionId, err)
		}
	}

	nextRunAt := time.Now().Add(time.Hour)
	matchingTask := &po.AutomationTask{
		Title:           "matching",
		UserId:          "user-1",
		ExecType:        po.AutomationExecTypeOnce,
		NextRunAt:       nextRunAt,
		Project:         "old-name",
		SharedProjectId: 0,
	}
	otherUserTask := &po.AutomationTask{
		Title:           "other-user",
		UserId:          "user-2",
		ExecType:        po.AutomationExecTypeOnce,
		NextRunAt:       nextRunAt,
		Project:         "old-name",
		SharedProjectId: 0,
	}
	sharedTask := &po.AutomationTask{
		Title:           "shared",
		UserId:          "user-1",
		ExecType:        po.AutomationExecTypeOnce,
		NextRunAt:       nextRunAt,
		Project:         "old-name",
		SharedProjectId: 7,
	}
	for _, task := range []*po.AutomationTask{matchingTask, otherUserTask, sharedTask} {
		if err := db.Create(task).Error; err != nil {
			t.Fatalf("create automation task %q: %v", task.Title, err)
		}
	}

	if err := repo.RenameProject(project, "new-name"); err != nil {
		t.Fatalf("rename project: %v", err)
	}

	var sessions []po.Session
	if err := db.Order("session_id").Find(&sessions).Error; err != nil {
		t.Fatalf("load sessions after rename: %v", err)
	}
	if sessions[0].Project != "new-name" || sessions[1].Project != "old-name" || sessions[2].Project != "old-name" {
		t.Fatalf("unexpected session projects after rename: %#v", sessions)
	}

	var tasks []po.AutomationTask
	if err := db.Order("id").Find(&tasks).Error; err != nil {
		t.Fatalf("load automation tasks after rename: %v", err)
	}
	if tasks[0].Project != "new-name" || tasks[1].Project != "old-name" || tasks[2].Project != "old-name" {
		t.Fatalf("unexpected automation task projects after rename: %#v", tasks)
	}

	if err := repo.ClearProjectReferences("user-1", "new-name"); err != nil {
		t.Fatalf("clear project references: %v", err)
	}

	if err := db.First(&matchingSession, "session_id = ?", matchingSession.SessionId).Error; err != nil {
		t.Fatalf("reload matching session: %v", err)
	}
	if matchingSession.Project != "" {
		t.Fatalf("matching session project = %q, want empty", matchingSession.Project)
	}
	if err := db.First(&matchingTask, matchingTask.Id).Error; err != nil {
		t.Fatalf("reload matching automation task: %v", err)
	}
	if matchingTask.Project != "" {
		t.Fatalf("matching automation task project = %q, want empty", matchingTask.Project)
	}
}

func TestUserProjectRepositoryRejectsStaleOrMissingRename(t *testing.T) {
	db, repo := newUserProjectRepositoryTest(t)

	project := &po.UserProject{UserId: "user-1", ProjectName: "current-name"}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}

	staleSession := &po.Session{
		SessionId: "stale-session",
		UserId:    "user-1",
		Project:   "old-name",
	}
	staleTask := &po.AutomationTask{
		Title:     "stale-task",
		UserId:    "user-1",
		ExecType:  po.AutomationExecTypeOnce,
		NextRunAt: time.Now().Add(time.Hour),
		Project:   "old-name",
	}
	missingSession := &po.Session{
		SessionId: "missing-session",
		UserId:    "user-2",
		Project:   "old-name",
	}
	missingTask := &po.AutomationTask{
		Title:     "missing-task",
		UserId:    "user-2",
		ExecType:  po.AutomationExecTypeOnce,
		NextRunAt: time.Now().Add(time.Hour),
		Project:   "old-name",
	}
	for _, session := range []*po.Session{staleSession, missingSession} {
		if err := db.Create(session).Error; err != nil {
			t.Fatalf("create session %q: %v", session.SessionId, err)
		}
	}
	for _, task := range []*po.AutomationTask{staleTask, missingTask} {
		if err := db.Create(task).Error; err != nil {
			t.Fatalf("create automation task %q: %v", task.Title, err)
		}
	}

	staleProject := &po.UserProject{Id: project.Id, UserId: "user-1", ProjectName: "old-name"}
	if err := repo.RenameProject(staleProject, "renamed"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("stale rename error = %v, want gorm.ErrRecordNotFound", err)
	}

	wrongOwnerProject := &po.UserProject{Id: project.Id, UserId: "user-2", ProjectName: "current-name"}
	if err := repo.RenameProject(wrongOwnerProject, "renamed"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("wrong-owner rename error = %v, want gorm.ErrRecordNotFound", err)
	}

	missingProject := &po.UserProject{Id: project.Id + 1, UserId: "user-2", ProjectName: "old-name"}
	if err := repo.RenameProject(missingProject, "renamed"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing rename error = %v, want gorm.ErrRecordNotFound", err)
	}

	var persistedProject po.UserProject
	if err := db.First(&persistedProject, project.Id).Error; err != nil {
		t.Fatalf("reload project: %v", err)
	}
	if persistedProject.ProjectName != "current-name" {
		t.Fatalf("project name after rejected renames = %q, want %q", persistedProject.ProjectName, "current-name")
	}

	var sessions []po.Session
	if err := db.Order("session_id").Find(&sessions).Error; err != nil {
		t.Fatalf("reload sessions: %v", err)
	}
	for _, session := range sessions {
		if session.Project != "old-name" {
			t.Fatalf("session %q project after rejected renames = %q, want %q", session.SessionId, session.Project, "old-name")
		}
	}

	var tasks []po.AutomationTask
	if err := db.Order("id").Find(&tasks).Error; err != nil {
		t.Fatalf("reload automation tasks: %v", err)
	}
	for _, task := range tasks {
		if task.Project != "old-name" {
			t.Fatalf("automation task %q project after rejected renames = %q, want %q", task.Title, task.Project, "old-name")
		}
	}
}

func TestUserProjectRepositoryUpdatesReferenceTimestamps(t *testing.T) {
	db, repo := newUserProjectRepositoryTest(t)

	project := &po.UserProject{UserId: "user-1", ProjectName: "old-name"}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	session := &po.Session{SessionId: "timestamp-session", UserId: "user-1", Project: "old-name"}
	task := &po.AutomationTask{
		Title:     "timestamp-task",
		UserId:    "user-1",
		ExecType:  po.AutomationExecTypeOnce,
		NextRunAt: time.Now().Add(time.Hour),
		Project:   "old-name",
	}
	if err := db.Create(session).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create automation task: %v", err)
	}

	renameBaseline := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := db.Model(&po.Session{}).
		Where("session_id = ?", session.SessionId).
		UpdateColumn("updated", renameBaseline).Error; err != nil {
		t.Fatalf("set session rename baseline: %v", err)
	}
	if err := db.Model(&po.AutomationTask{}).
		Where("id = ?", task.Id).
		UpdateColumn("updated", renameBaseline).Error; err != nil {
		t.Fatalf("set automation task rename baseline: %v", err)
	}

	var beforeSession po.Session
	if err := db.First(&beforeSession, "session_id = ?", session.SessionId).Error; err != nil {
		t.Fatalf("load session before rename: %v", err)
	}
	var beforeTask po.AutomationTask
	if err := db.First(&beforeTask, task.Id).Error; err != nil {
		t.Fatalf("load automation task before rename: %v", err)
	}
	if !beforeSession.Updated.Equal(renameBaseline) || !beforeTask.Updated.Equal(renameBaseline) {
		t.Fatalf("test timestamp baseline not applied: session=%v task=%v", beforeSession.Updated, beforeTask.Updated)
	}
	if err := repo.RenameProject(project, "new-name"); err != nil {
		t.Fatalf("rename project: %v", err)
	}

	var renamedSession po.Session
	if err := db.First(&renamedSession, "session_id = ?", session.SessionId).Error; err != nil {
		t.Fatalf("load session after rename: %v", err)
	}
	var renamedTask po.AutomationTask
	if err := db.First(&renamedTask, task.Id).Error; err != nil {
		t.Fatalf("load automation task after rename: %v", err)
	}
	if !renamedSession.Updated.After(renameBaseline) {
		t.Fatalf("session updated = %v, want after %v", renamedSession.Updated, renameBaseline)
	}
	if !renamedTask.Updated.After(renameBaseline) {
		t.Fatalf("automation task updated = %v, want after %v", renamedTask.Updated, renameBaseline)
	}
	if !renamedSession.Updated.Equal(renamedTask.Updated) {
		t.Fatalf("rename timestamps differ: session=%v task=%v", renamedSession.Updated, renamedTask.Updated)
	}

	clearBaseline := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := db.Model(&po.Session{}).
		Where("session_id = ?", session.SessionId).
		UpdateColumn("updated", clearBaseline).Error; err != nil {
		t.Fatalf("set session clear baseline: %v", err)
	}
	if err := db.Model(&po.AutomationTask{}).
		Where("id = ?", task.Id).
		UpdateColumn("updated", clearBaseline).Error; err != nil {
		t.Fatalf("set automation task clear baseline: %v", err)
	}
	if err := repo.ClearProjectReferences("user-1", "new-name"); err != nil {
		t.Fatalf("clear project references: %v", err)
	}
	if err := db.First(&renamedSession, "session_id = ?", session.SessionId).Error; err != nil {
		t.Fatalf("load session after clear: %v", err)
	}
	if err := db.First(&renamedTask, task.Id).Error; err != nil {
		t.Fatalf("load automation task after clear: %v", err)
	}
	if !renamedSession.Updated.After(clearBaseline) {
		t.Fatalf("session updated after clear = %v, want after %v", renamedSession.Updated, clearBaseline)
	}
	if !renamedTask.Updated.After(clearBaseline) {
		t.Fatalf("automation task updated after clear = %v, want after %v", renamedTask.Updated, clearBaseline)
	}
	if !renamedSession.Updated.Equal(renamedTask.Updated) {
		t.Fatalf("clear timestamps differ: session=%v task=%v", renamedSession.Updated, renamedTask.Updated)
	}
}

func TestUserProjectRepositoryCRUDAndUsernameLookup(t *testing.T) {
	db, repo := newUserProjectRepositoryTest(t)

	project := &po.UserProject{
		UserId:      "user-1",
		ProjectName: "crud-project",
		Description: "before",
		GitUrl:      "https://example.com/repository.git",
	}
	if err := repo.Create(project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	byID, err := repo.GetByID(project.Id)
	if err != nil {
		t.Fatalf("get project by ID: %v", err)
	}
	if byID.GitUrl != project.GitUrl {
		t.Fatalf("GitUrl = %q, want %q", byID.GitUrl, project.GitUrl)
	}
	byName, err := repo.GetByName(project.UserId, project.ProjectName)
	if err != nil {
		t.Fatalf("get project by name: %v", err)
	}
	if byName.Id != project.Id {
		t.Fatalf("project by name ID = %d, want %d", byName.Id, project.Id)
	}
	if err := repo.UpdateDescription(project.Id, "after"); err != nil {
		t.Fatalf("update project description: %v", err)
	}
	byID, err = repo.GetByID(project.Id)
	if err != nil {
		t.Fatalf("get updated project: %v", err)
	}
	if byID.Description != "after" {
		t.Fatalf("description = %q, want %q", byID.Description, "after")
	}

	user := &po.User{
		UserId:   "deleted-user",
		Username: "retained-username",
		Password: "password-hash",
		Deleted:  1,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create deleted user: %v", err)
	}
	username, err := repo.GetUsernameByUserID(user.UserId)
	if err != nil {
		t.Fatalf("get username for deleted user: %v", err)
	}
	if username != user.Username {
		t.Fatalf("username = %q, want %q", username, user.Username)
	}

	if err := repo.Delete(project.Id); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := repo.GetByID(project.Id); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("get deleted project error = %v, want gorm.ErrRecordNotFound", err)
	}
}

func newUserProjectRepositoryTest(t *testing.T) (*gorm.DB, *repository.UserProjectRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&po.UserProject{}, &po.Session{}, &po.AutomationTask{}, &po.User{}); err != nil {
		t.Fatalf("migrate project reference tables: %v", err)
	}

	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} {
		return db
	})
	unitTest.Run()

	var repo *repository.UserProjectRepository
	unitTest.FetchRepository(&repo)
	return db, repo
}
