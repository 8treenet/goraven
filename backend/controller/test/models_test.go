package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestGetProviders(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/providers").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetRecommendModels(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/providers/recommend?providerId=open_router&apiKey=sk-").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetModels(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetModelsByProvider(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models?providerId=deepseek").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetModelsBySearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models?search=chat").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetModelDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models/2").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateModel(t *testing.T) {
	req := vo.AdminCreateModelReq{
		ProviderDisplayName: "火山3",
		ProviderID:          "openai_compatible",
		ModelName:           "deepseek-v3-1-terminus",
		APIKey:              "ddd-8a36-4dfd-a026-bed断点c",
		BaseURL:             "https://ark.cn-beijing.volces.com/api/v3",
		ExtraFields:         "{\"thinking\":{\"type\":\"enabled\"}}",

		ContextLen: 200,
		IsDefault:  0,
		IsFlash:    0,
		Remark:     "test model",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdateModel(t *testing.T) {
	req := vo.AdminUpdateModelReq{

		ProxyURL: "http://127.0.0.1:7898",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models/1").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestUpdateModelStatus(t *testing.T) {
	req := map[string]interface{}{
		"status": 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models/1/status").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestSetDefaultModel(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models/2/default").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestSetFlashModel(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models/1/flash").
		Put().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeleteModel(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/models/2").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
