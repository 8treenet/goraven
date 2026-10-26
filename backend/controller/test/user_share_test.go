package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// TestListMySharesWithPage 分页查询分享列表 GET /api/sessions/shares?page=1&pageSize=5
func TestListMySharesWithPage(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/shares?page=1&pageSize=5").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreateShare 创建会话分享链接 POST /api/sessions/:sessionId/share
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

// TestCreateShareWithTitle 创建带自定义标题的分享链接
func TestCreateShareWithTitle(t *testing.T) {
	req := vo.CreateShareReq{
		Title:     "自定义分享标题",
		ExpiresIn: "7d",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/3d1cd0291e804fa5bdfc201b0cae407a/share").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetSessionShare 获取会话的分享链接 GET /api/sessions/:sessionId/share
func TestGetSessionShare(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/af850c58f2b845078924f5bb6c795be7/share").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetSessionShareNotFound 查询不存在分享的会话
func TestGetSessionShareNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/nonexistent/share").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestDeleteShare 删除会话的分享链接 DELETE /api/sessions/:sessionId/share
func TestDeleteShare(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/3d1cd0291e804fa5bdfc201b0cae407a/share").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetPublicShare 公开访问分享页面 GET /api/share/:shareId（无鉴权）
func TestGetPublicShare(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain + "/api/share/3693b9431f7442d89e4b2b0d53cfcd81").
		Get().
		ToString()
	t.Log(str, httpResp)
}

// TestGetPublicShareNotFound 公开访问不存在的分享（无鉴权）
func TestGetPublicShareNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain + "/api/share/a555e97d665b44d5b77c06af2ecca9d9").
		Get().
		ToString()
	t.Log(str, httpResp)
}
