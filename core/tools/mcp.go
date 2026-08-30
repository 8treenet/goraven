package tools

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	rconf "goraven/config"
	"goraven/core/sandbox"
	"goraven/util"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
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
	jsonMap := map[string]string{}
	if obj.HttpHeader != "" {
		mcpErr = json.Unmarshal([]byte(obj.HttpHeader), &jsonMap)
	}
	if mcpErr != nil {
		return
	}
	switch obj.Transport {
	case MCPTransportSSE:
		var options []transport.ClientOption
		if len(jsonMap) > 0 {
			options = append(options, transport.WithHeaders(jsonMap))
		}
		if obj.ProxyURL != "" {
			options = append(options, transport.WithHTTPClient(util.GetProxyHTTPClient(obj.ProxyURL)))
		}
		client, mcpErr = mcpclient.NewSSEMCPClient(obj.HttpURL, options...)
	case MCPTransportStdio:
		if obj.StdioType == MCPStdioNPX {
			obj.StdioArgs = append([]string{"-y"}, obj.StdioArgs...)
			obj.StdioCommand = "npx"
		} else {
			obj.StdioCommand = "uvx"
		}

		client, mcpErr = box.NewStdioMCPClient(obj.StdioCommand, obj.StdioENV, obj.StdioArgs...)
	case MCPTransportHTTP:
		var options []transport.StreamableHTTPCOption
		if len(jsonMap) > 0 {
			options = append(options, transport.WithHTTPHeaders(jsonMap))
		}
		options = append(options, transport.WithHTTPTimeout(time.Duration(rconf.Get().Tools.HTTPTimeoutSeconds)*time.Second))
		if obj.ProxyURL != "" {
			options = append(options, transport.WithHTTPBasicClient(util.GetProxyHTTPClient(obj.ProxyURL)))
		}
		client, mcpErr = mcpclient.NewStreamableHttpClient(obj.HttpURL, options...)
	}
	return
}

func ValidateMCP(ctx context.Context, obj MCP, box sandbox.Sandbox) (toolNames []string, err error) {
	client, err := GetMCPClient(obj, box)
	if err != nil {
		return
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err = client.Initialize(ctx, initRequest); err != nil {
		return
	}

	mcpTools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: client})
	if err != nil {
		return
	}

	for _, t := range mcpTools {
		toolInfo, infoErr := t.Info(ctx)
		if infoErr != nil {
			continue
		}
		toolNames = append(toolNames, toolInfo.Name)
	}

	if len(toolNames) == 0 {
		err = ErrNoToolsFound
	}
	return
}

var ErrNoToolsFound = errors.New("no tools found in MCP server")
