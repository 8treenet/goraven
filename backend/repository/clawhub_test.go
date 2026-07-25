package repository_test

import (
	"goraven/backend/repository"
	"goraven/config"
	unit_test "goraven/util/unit"
	"testing"
)

func TestClawHubRepository_Search(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()
	t.Log(config.Get().GetClawHUBCacheDir())
	return

	var repo *repository.ClawHubRepository
	unitTest.FetchRepository(&repo)
	repo.ClawHubToken = "clh_RMufsbRbX2g2xCRBnZRzVX3arstiBvuMmLNzN18m85U"

	data, err := repo.Search("Bloomberg", 10)
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}

func TestClawHubRepository_Explore(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.ClawHubRepository
	unitTest.FetchRepository(&repo)
	repo.ClawHubToken = "clh_RMufsbRbX2g2xCRBnZRzVX3arstiBvuMmLNzN18m85U"

	data, err := repo.Explore("stars")
	if err != nil {
		panic(err)
	}
	t.Log(len(data.Items))
	t.Log(unit_test.JsonLog(data))
}

func TestClawHubRepository_FetchFile(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.ClawHubRepository
	unitTest.FetchRepository(&repo)
	repo.ClawHubToken = "clh_RMufsbRbX2g2xCRBnZRzVX3arstiBvuMmLNzN18m85U"

	data, err := repo.FetchFile("bloomberg-headlines", "SKILL.md")
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}

func TestClawHubRepository_Download(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.ClawHubRepository
	unitTest.FetchRepository(&repo)
	t.Log(repo.Download("bloomberg-headlines"))

}

func TestClawHubRepository_UnzipFile(t *testing.T) {

}
