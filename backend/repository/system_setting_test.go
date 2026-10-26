package repository_test

import (
	"context"
	"goraven/backend/repository"
	unit_test "goraven/util/unit"
	"testing"
	"time"
)

func TestSystemSettingRepository_LoadConfig(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.SystemSettingRepository
	unitTest.FetchRepository(&repo)

	t.Log(repo.LoadConfig())
	t.Log(repo.LoadConfig())
	t.Log(repo.Update("sharing.file_expires_hours", "73"))
	t.Log(repo.LoadConfig())
}

func TestCache(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.SystemSettingRepository
	unitTest.FetchRepository(&repo)
	repo.Redis().SetNX(context.Background(), "test-key", "123", time.Second*20)
	repo.Redis().SetNX(context.Background(), "test-timeout", "456", time.Second*3)
	t.Log(repo.Redis().Get(context.Background(), "test-key").Result())
	time.Sleep(5 * time.Second)
	t.Log(repo.Redis().Get(context.Background(), "test-timeout").Result())
}
