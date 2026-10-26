package test

import (
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// TestGetMCPs MCP 列表 GET /api/admin/mcp
func TestGetMCPs(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMCPsWithSearch MCP 列表带搜索
func TestGetMCPsWithSearch(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp?search=Web").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMCPsWithTransportFilter MCP 列表按传输类型筛选
func TestGetMCPsWithTransportFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp?transport=Stdio").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMCPsWithStatusFilter MCP 列表按状态筛选
func TestGetMCPsWithStatusFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp?status=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetMCPDetail MCP 详情 GET /api/admin/mcp/:id
func TestGetMCPDetail(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp/3").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreateMCPStdio 创建 Stdio/npx 类型 MCP
func TestCreateMCPStdio(t *testing.T) {
	req := vo.AdminCreateMCPReq{
		Name:        "sequential-thinking-test",
		DisplayName: "Sequential Thinking Test",
		Icon:        "🧠",
		Description: "顺序思考和任务拆解（测试）",
		Transport:   "Stdio",
		StdioType:   "npx",
		StdioArgs:   []string{"@modelcontextprotocol/server-sequential-thinking"},
		StdioEnv:    map[string]string{},
		Remark:      "test mcp",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreateMCPStdioUvx 创建 Stdio/uvx 类型 MCP
func TestCreateMCPStdioUvx(t *testing.T) {
	req := vo.AdminCreateMCPReq{
		Name:        "mcp-server-fetch-test",
		DisplayName: "MCP Server Fetch Test",
		Icon:        "🌐",
		Description: "网页抓取工具（测试）",
		Transport:   "Stdio",
		StdioType:   "uvx",
		StdioArgs:   []string{"mcp-server-fetch"},
		StdioEnv:    map[string]string{},
		Remark:      "test mcp uvx",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreateMCPSSE 创建 SSE 类型 MCP
func TestCreateMCPSSE(t *testing.T) {
	req := vo.AdminCreateMCPReq{
		Name:         "sse-mcp-test",
		DisplayName:  "SSE MCP Test",
		Icon:         "📡",
		Description:  "SSE 传输类型测试",
		Transport:    "SSE",
		HttpUrl:      "https://mcp.example.com/sse",
		HttpHeader:   map[string]string{"Authorization": "Bearer test-token"},
		HttpProxyUrl: "http://127.0.0.1:7890",
		Remark:       "test sse mcp",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreateMCPHTTP 创建 StreamableHttp 类型 MCP
func TestCreateMCPHTTP(t *testing.T) {
	req := vo.AdminCreateMCPReq{
		Name:         "web-search1",
		DisplayName:  "WebSearch2",
		Icon:         "🔗",
		Description:  "HTTP 传输类型测试",
		Transport:    "StreamableHttp",
		HttpUrl:      "https://dashscope.aliyuncs.com/api/v1/mcps/WebSearch/mcp",
		HttpHeader:   map[string]string{"Authorization": "Bearer sk-ba11bbf24d034b94b8e47b38d95f158c"},
		HttpProxyUrl: "",
		Remark:       "test http mcp",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCreateMCPWithEnv 创建带环境变量的 Stdio MCP
func TestCreateMCPWithEnv(t *testing.T) {
	req := vo.AdminCreateMCPReq{
		Name:        "github-mcp-test",
		DisplayName: "GitHub MCP Test",
		Icon:        "🐙",
		Description: "GitHub 操作工具（测试）",
		Transport:   "Stdio",
		StdioType:   "npx",
		StdioArgs:   []string{"@modelcontextprotocol/server-github"},
		StdioEnv: map[string]string{
			"GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_test_token_xxx",
		},
		Remark: "test mcp with env",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdateMCP 编辑 MCP PUT /api/admin/mcp/:id
func TestUpdateMCP(t *testing.T) {
	req := vo.AdminUpdateMCPReq{
		DisplayName:  "WebSearch234",
		Icon:         "🎯",
		Description:  "更新后的描述",
		Remark:       "updated",
		HttpProxyUrl: "http://127.0.0.1:7897",
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp/3").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestUpdateMCPStatus 启用/禁用 MCP PUT /api/admin/mcp/:id/status
func TestUpdateMCPStatus(t *testing.T) {
	req := map[string]interface{}{
		"status": 0,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp/3/status").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestEnableMCP 启用 MCP
func TestEnableMCP(t *testing.T) {
	req := map[string]interface{}{
		"status": 1,
	}
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp/1/status").
		Put().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestDeleteMCP 删除 MCP DELETE /api/admin/mcp/:id
func TestDeleteMCP(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp/4").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestCheckMCPHealth 手动触发 MCP 健康检查 POST /api/admin/mcp/healthCheck
func TestCheckMCPHealth(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp/healthCheck").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

// TestGetRecommendMCPs 推荐 MCP 列表 GET /api/admin/mcp/recommend
func TestGetRecommendMCPs(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/admin/mcp/recommend").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
