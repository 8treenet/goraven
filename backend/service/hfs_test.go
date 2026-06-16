package service_test

import (
	"raven/backend/service"
	"raven/config"
	unit_test "raven/util/unit"
	"testing"
)

func TestHFSService_GenerateURL(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *service.HFSService
	unitTest.FetchService(&service)

	t.Log(service.GenerateURL("999", "/documents/go_concurrency_plan_20260425_82731.md"))
}

func TestWriteConf(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	t.Log(config.Get().ModifyConfig("system", "initialized", "false"))
}
