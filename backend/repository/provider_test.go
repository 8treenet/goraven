package repository_test

import (
	"goraven/backend/repository"
	unit_test "goraven/util/unit"
	"testing"
)

func TestAIModelRepository_GetFlashChatModel(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.ProviderRepository
	unitTest.FetchRepository(&repo)
	t.Log(repo.GetFlashChatModel(33))
}

func TestProviderRepository_CreateChatModelFromID(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.ProviderRepository
	unitTest.FetchRepository(&repo)
	t.Log(repo.GetFlashChatModel(1))
	t.Log(repo.CreateChatModelFromID(1, true))
}
