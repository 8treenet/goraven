package service_test

import (
	"raven/backend/service"
	unit_test "raven/util/unit"
	"testing"
)

func TestAIModelService_RecommendModels(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *service.AIModelService
	unitTest.FetchService(&service)

	t.Log(service.RecommendModels("deepseek", "sk-mock", ""))
}
