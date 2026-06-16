package service_test

import (
	"raven/backend/service"
	"raven/backend/vo"
	unit_test "raven/util/unit"
	"testing"
	"time"
)

func TestSkillService_InstallSkill(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *service.SkillService
	unitTest.FetchService(&service)
	resp, err := service.InstallSkill("999", &vo.UserSkillInstallReq{
		SkillId: 7,
	})
	if err != nil {
		panic(err)
	}
	t.Log(resp)
	time.Sleep(160 * time.Second)
}

func TestSkillService_RetryInstallSkill(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *service.SkillService
	unitTest.FetchService(&service)

	err := service.RetryInstallSkill(9, "999")
	if err != nil {
		panic(err)
	}
	time.Sleep(140 * time.Second)
}
