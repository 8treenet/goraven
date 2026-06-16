package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestListMySharesWithPage(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/shares?page=1&pageSize=5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateShare(t *testing.T) {
	req := vo.CreateShareReq{
		ExpiresIn: "24h",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/af850c58f2b845078924f5bb6c795be7/share").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestCreateShareWithTitle(t *testing.T) {
	req := vo.CreateShareReq{
		Title:		"自定义分享标题",
		ExpiresIn:	"7d",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/3d1cd0291e804fa5bdfc201b0cae407a/share").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetSessionShare(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/af850c58f2b845078924f5bb6c795be7/share").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetSessionShareNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/nonexistent/share").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeleteShare(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/3d1cd0291e804fa5bdfc201b0cae407a/share").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetPublicShare(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain + "/api/share/3693b9431f7442d89e4b2b0d53cfcd81").
		Get().
		ToString()
	t.Log(str, httpResp)
}

func TestGetPublicShareNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain + "/api/share/a555e97d665b44d5b77c06af2ecca9d9").
		Get().
		ToString()
	t.Log(str, httpResp)
}
