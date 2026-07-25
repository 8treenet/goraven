package service_test

import (
	"goraven/backend/service"
	unit_test "goraven/util/unit"
	"testing"
)

func TestAIModelService_RecommendModels(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *service.AIModelService
	unitTest.FetchService(&service)

	t.Log(service.RecommendModels("deepseek", "sk-ccec93010d58433699b4abda77bc372f", ""))
}
