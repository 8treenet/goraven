package repository_test

import (
	"raven/backend/repository"
	unit_test "raven/util/unit"
	"testing"
)

func TestAIModelRepository_GetCompressChatModel(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.ProviderRepository
	unitTest.FetchRepository(&repo)
	t.Log(repo.GetCompressChatModel(33))
}

func TestProviderRepository_CreateChatModelFromID(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.ProviderRepository
	unitTest.FetchRepository(&repo)
	t.Log(repo.GetCompressChatModel(1))
	t.Log(repo.CreateChatModelFromID(1, true))
}
