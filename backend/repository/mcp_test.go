package repository_test

import (
	"raven/backend/repository"
	unit_test "raven/util/unit"
	"testing"
)

func TestMCPRepository_GetMCPEndpointsByIDs(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MCPRepository
	unitTest.FetchRepository(&repo)

	t.Log(repo.GetMCPEndpointsByIDs([]int{3, 4, 5}))
}
