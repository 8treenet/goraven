package tools

import (
	"context"

	"goraven/core/sandbox"

	mcpclient "github.com/mark3labs/mcp-go/client"
)

const (
	MCPTransportStdio = "Stdio"
	MCPTransportSSE   = "SSE"
	MCPTransportHTTP  = "StreamableHttp"
	MCPStdioNPX       = "npx"
	MCPStdioUVX       = "uvx"
)

type MCP struct {
	Transport   string
	Name        string
	DisplayName string
	HttpURL     string
	HttpHeader  string
	ProxyURL    string

	StdioType    string
	StdioCommand string
	StdioENV     []string
	StdioArgs    []string
}

func GetMCPClient(obj MCP, box sandbox.Sandbox) (client *mcpclient.Client, mcpErr error) {
	return nil, nil
}

func ValidateMCP(ctx context.Context, obj MCP, box sandbox.Sandbox) (toolNames []string, err error) {
	return
}
