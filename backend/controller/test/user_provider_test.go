package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestGetUserModels(t *testing.T) {
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	[]vo.UserModelItem	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/providers/models").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("GetUserModels error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("GetUserModels failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("GetUserModels success: count=%d", len(rsp.Data))
	for _, m := range rsp.Data {
		t.Logf("  model: aiModelId=%d providerDisplayName=%s modelName=%s contextLen=%d isDefault=%d",
			m.AIModelId, m.ProviderDisplayName, m.ModelName, m.ContextLen, m.IsDefault)
	}
}

func TestGetUserMCPs(t *testing.T) {
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	[]vo.UserMCPItem	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/mcp").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("GetUserMCPs error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("GetUserMCPs failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("GetUserMCPs success: count=%d", len(rsp.Data))
	for _, m := range rsp.Data {
		t.Logf("  mcp: mcpId=%d name=%s icon=%s description=%s",
			m.McpId, m.Name, m.Icon, m.Description)
	}
}

func TestGetSimpleSkills(t *testing.T) {
	var rsp struct {
		Code	int				`json:"code"`
		Msg	string				`json:"msg"`
		Data	[]vo.UserAvailableSkillItem	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/skills/simpleSkills").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("GetSimpleSkills error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("GetSimpleSkills failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("GetSimpleSkills success: count=%d", len(rsp.Data))
	for _, s := range rsp.Data {
		t.Logf("  skill: userSkillId=%d skillName=%s description=%s icon=%s source=%s categoryId=%d categoryName=%s",
			s.UserSkillId, s.SkillName, s.Description, s.Icon, s.Source, s.CategoryId, s.CategoryName)
	}
}
