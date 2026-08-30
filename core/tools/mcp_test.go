package tools_test

import (
	"context"
	"encoding/json"
	"goraven/core/sandbox"
	"goraven/core/tools"
	"testing"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetMCPClient(t *testing.T) {
	obj := tools.MCP{
		Transport:    tools.MCPTransportHTTP,
		Name:         "WebSearch",
		HttpURL:      "https://dashscope.aliyuncs.com/api/v1/mcps/WebSearch/mcp",
		HttpHeader:   `{"Authorization":"Bearer "}`,
		StdioCommand: "",
		StdioENV:     []string{},
		StdioArgs:    []string{},
	}

	box, _ := sandbox.NewSandbox("999")
	client, mcpErr := tools.GetMCPClient(obj, box)
	if mcpErr != nil {
		panic(mcpErr)
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	_, err := client.Initialize(context.Background(), initRequest)
	if err != nil {
		return
	}

	mcpTools, e := mcpp.GetTools(context.Background(), &mcpp.Config{Cli: client})
	if e != nil {
		err = e
		return
	}
	for _, v := range mcpTools {
		toolobj, infoerr := v.Info(context.Background())
		if infoerr != nil {
			panic(infoerr)
		}
		schema, _ := toolobj.ToJSONSchema()
		jdata, _ := json.Marshal(schema)
		t.Logf("工具名: %s, 描述: %s, schema:%v", toolobj.Name, toolobj.Desc, string(jdata))

		invokable, ok := v.(tool.InvokableTool)
		if !ok {
			t.Logf("工具不支持直接调用")
			continue
		}
		result, runerr := invokable.InvokableRun(context.Background(), `{"query":"美国和伊朗的现在如何了？"}`)
		if runerr != nil {
			t.Logf("调用失败: %v", runerr)
			continue
		}
		t.Logf("调用结果: %s", result)
	}
}

func TestValidateMCP(t *testing.T) {
	obj := tools.MCP{
		Transport:    tools.MCPTransportHTTP,
		Name:         "WebSearch",
		HttpURL:      "https://dashscope.aliyuncs.com/api/v1/mcps/WebSearch/mcp",
		HttpHeader:   `{"Authorization":"Bearer "}`,
		StdioCommand: "",
		StdioENV:     []string{},
		StdioArgs:    []string{},
	}
	box, _ := sandbox.NewSandbox("999")
	t.Log(tools.ValidateMCP(context.Background(), obj, box))
}

func TestNPXMCPClient(t *testing.T) {
	obj := tools.MCP{
		Transport: tools.MCPTransportStdio,
		StdioType: tools.MCPStdioNPX,
		//StdioType: tools.MCPStdioUVX,
		Name:      "dddd",
		StdioArgs: []string{"@theo.foobar/mcp-time"},
		//StdioArgs: []string{"mcp-server-time"},
	}
	box, _ := sandbox.NewSandbox("999")
	client, mcpErr := tools.GetMCPClient(obj, box)
	if mcpErr != nil {
		panic(mcpErr)
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	_, err := client.Initialize(context.Background(), initRequest)
	if err != nil {
		return
	}

	mcpTools, e := mcpp.GetTools(context.Background(), &mcpp.Config{Cli: client})
	if e != nil {
		err = e
		return
	}
	for _, v := range mcpTools {
		toolobj, infoerr := v.Info(context.Background())
		if infoerr != nil {
			panic(infoerr)
		}
		schema, _ := toolobj.ToJSONSchema()
		jdata, _ := json.Marshal(schema)
		t.Logf("工具名: %s, 描述: %s, schema:%v", toolobj.Name, toolobj.Desc, string(jdata))

		invokable, ok := v.(tool.InvokableTool)
		if !ok {
			t.Logf("工具不支持直接调用")
			continue
		}
		var args string
		switch toolobj.Name {
		case "current_time":
			args = `{}`
		case "add_time":
			args = `{"duration":"1h"}`
		case "compare_time":
			args = `{"time_a":"2024-01-01","time_b":"2024-01-02"}`
		case "convert_timezone":
			args = `{}`
		case "relative_time":
			args = `{"text":"now"}`
		default:
			t.Logf("未知工具: %s, 跳过", toolobj.Name)
			continue
		}
		result, runerr := invokable.InvokableRun(context.Background(), args)
		if runerr != nil {
			t.Logf("调用失败: %v", runerr)
			continue
		}
		t.Logf("调用结果: %s", result)
	}
}

func validateMCPHelper(obj tools.MCP, t *testing.T) ([]string, bool) {
	box, _ := sandbox.NewSandbox("999")
	toolNames, err := tools.ValidateMCP(context.Background(), obj, box)
	if err != nil {
		t.Logf("[FAIL] %s: %v", obj.Name, err)
		return nil, false
	}
	t.Logf("[OK] %s: tools=%v", obj.Name, toolNames)
	return toolNames, true
}

func TestSeedMCPValidate(t *testing.T) {
	// These 4 MCPs correspond to IDs 13-16 in 02_models_and_mcp.sql
	// All use npx Stdio transport, no API keys required
	seedMCPs := []tools.MCP{
		{
			Transport: tools.MCPTransportStdio,
			StdioType: tools.MCPStdioNPX,
			Name:      "memory",
			StdioArgs: []string{"@modelcontextprotocol/server-memory"},
		},
		{
			Transport: tools.MCPTransportStdio,
			StdioType: tools.MCPStdioNPX,
			Name:      "sequential-thinking",
			StdioArgs: []string{"@modelcontextprotocol/server-sequential-thinking"},
		},
		{
			Transport: tools.MCPTransportStdio,
			StdioType: tools.MCPStdioNPX,
			Name:      "time-tool",
			StdioArgs: []string{"@theo.foobar/mcp-time"},
		},
		{
			Transport: tools.MCPTransportStdio,
			StdioType: tools.MCPStdioNPX,
			Name:      "mcp-everything",
			StdioArgs: []string{"@modelcontextprotocol/server-everything"},
		},
	}

	for _, c := range seedMCPs {
		if _, ok := validateMCPHelper(c, t); !ok {
			t.Fatalf("Seed MCP %s failed health check", c.Name)
		}
	}
}

func TestValidatebraveMCP(t *testing.T) {
	obj := tools.MCP{
		Transport: tools.MCPTransportStdio,
		Name:      "BraveSearch",
		StdioType: tools.MCPStdioNPX,
		StdioENV:  []string{"BRAVE_API_KEY=BSA"},
		StdioArgs: []string{"@brave/brave-search-mcp-server"},
	}
	box, _ := sandbox.NewSandbox("2ba0b9431cfb465c8c7571a4f03b7d0a")
	t.Log(tools.ValidateMCP(context.Background(), obj, box))
}
