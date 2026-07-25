package repository_test

import (
	"goraven/backend/repository"
	unit_test "goraven/util/unit"
	"testing"
)

func TestSkillRepository_GetUserSkillNamesByIDs(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.SkillRepository
	unitTest.FetchRepository(&repo)

	t.Log(repo.GetUserSkillNamesByIDs("999", []int{9, 10, 11}))
}
