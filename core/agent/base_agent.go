package agent

import (
	"context"
	"errors"
	"fmt"
	"goraven/backend/repository"
	"goraven/core/iface"
	"goraven/core/middleware"
	"goraven/core/plugin"
	"goraven/core/sandbox"
	"goraven/core/tools"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/adk"
	mfilesystem "github.com/cloudwego/eino/adk/middlewares/filesystem"
	"github.com/cloudwego/eino/adk/middlewares/patchtoolcalls"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewBaseAgent(name, instruction, description string, chatmodel iface.BaseChatModel, sysCfg *repository.SystemConfig, sandbox sandbox.Sandbox) *BaseAgent {
	maxIterations := sysCfg.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 150
	}
	conf := &adk.ChatModelAgentConfig{
		Name:          name,
		Description:   description,
		Instruction:   instruction,
		Model:         chatmodel,
		MaxIterations: maxIterations,
	}

	return &BaseAgent{
		conf:           conf,
		chatModel:      chatmodel,
		sysCfg:         sysCfg,
		sandbox:        sandbox,
		mcpToolDisplay: map[string]tools.MCP{},
		toolNames:      map[string]struct{}{},
	}
}

type BaseAgent struct {
	chatModel   iface.BaseChatModel
	conf        *adk.ChatModelAgentConfig
	sysCfg      *repository.SystemConfig
	sandbox     sandbox.Sandbox
	toolList    []tool.BaseTool
	toolNames   map[string]struct{}
	middlewares []adk.ChatModelAgentMiddleware

	mcpList             []tools.MCP
	mcpToolDisplay      map[string]tools.MCP
	skillAccessor       iface.SkillAccessor
	mcpFilter           iface.MCPFilter
	systemSkillProvider iface.SystemSkillProvider
	validateCommandFunc func(string) error
	Plugins             *plugin.Plugins
}

// SetSkillAccessor
func (base *BaseAgent) SetSkillAccessor(filter iface.SkillAccessor) {
	base.skillAccessor = filter
}

// SetSkillFilter
func (base *BaseAgent) SetMCPFilter(filter iface.MCPFilter) {
	base.mcpFilter = filter
}

// SetSystemSkillProvider 设置系统技能提供者
func (base *BaseAgent) SetSystemSkillProvider(provider iface.SystemSkillProvider) {
	base.systemSkillProvider = provider
}

// SetValidateCommand 设置安全审核函数
// 传入 nil 使用默认的严格模式 (ValidateCommand)
func (base *BaseAgent) SetValidateCommand(fn func(string) error) {
	if fn == nil {
		fn = ValidateCommand
	}
	base.validateCommandFunc = fn
}

// AddTool
func (base *BaseAgent) AddTool(tool tool.BaseTool) {
	base.toolList = append(base.toolList, tool)
}

// AddMCP
func (base *BaseAgent) AddMCP(mcp tools.MCP) {
	base.mcpList = append(base.mcpList, mcp)
}

// AddChatModelAgentMiddleware
func (base *BaseAgent) AddChatModelAgentMiddleware(middleware adk.ChatModelAgentMiddleware) {
	base.middlewares = append(base.middlewares, middleware)
}

// SetEmitInternalEvents enables or disables emitting internal events from agent tools.
// When enabled, events from sub-agents (like sub_agent) will be streamed to the parent agent.
func (base *BaseAgent) SetEmitInternalEvents(enabled bool) {
	base.conf.ToolsConfig.EmitInternalEvents = enabled
}

// SetModelRetryConfig 设置模型重试配置
func (base *BaseAgent) SetModelRetryConfig(config *adk.ModelRetryConfig) {
	base.conf.ModelRetryConfig = config
}

// defaultModelRetryConfig 返回默认的模型重试配置
func (base *BaseAgent) defaultModelRetryConfig() *adk.ModelRetryConfig {
	return &adk.ModelRetryConfig{
		MaxRetries: 3,
		BackoffFunc: func(ctx context.Context, attempt int) time.Duration {
			return time.Duration(attempt) * 3 * time.Second
		},
		IsRetryAble: func(ctx context.Context, err error) bool {
			if err == nil {
				return false
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return false
			}
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "too many requests") {
				time.Sleep(10 * time.Second)
			}
			return true
		},
	}
}

// addTools appends tools to the config, skipping any whose name already exists.
func (base *BaseAgent) addTools(newTools ...tool.BaseTool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, t := range newTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		base.addToolWithName(info.Name, t)
	}
}

// addToolWithName appends a tool with a known name, skipping duplicates.
func (base *BaseAgent) addToolWithName(name string, t tool.BaseTool) {
	if _, dup := base.toolNames[name]; dup {
		return
	}
	base.conf.ToolsConfig.Tools = append(base.conf.ToolsConfig.Tools, t)
	base.toolNames[name] = struct{}{}
}

func (base *BaseAgent) addMCPToolDisplay(toolName string, mcp tools.MCP) {
	base.mcpToolDisplay[toolName] = mcp
}

func (base *BaseAgent) GetMCPToolDisplay(toolName string) (tools.MCP, bool) {
	mcp, ok := base.mcpToolDisplay[toolName]
	return mcp, ok
}

func (base *BaseAgent) loadMCPTools(ctx context.Context) error {
	for _, v := range base.mcpList {
		if base.mcpFilter != nil && !base.mcpFilter.IsMCPSelected(v.Name) {
			//如果设置了mcpFilter，并且未勾选
			continue
		}
		client, mcpErr := tools.GetMCPClient(v, base.sandbox)
		if mcpErr != nil {
			return mcpErr
		}

		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		if _, err := client.Initialize(ctx, initRequest); err != nil {
			return err
		}

		mcpTools, e := mcpp.GetTools(ctx, &mcpp.Config{Cli: client})
		if e != nil {
			return e
		}

		for _, t := range mcpTools {
			toolInfo, infoErr := t.Info(ctx)
			if infoErr != nil {
				continue
			}
			base.addMCPToolDisplay(toolInfo.Name, v)

			// Only wrap tools whose schema has non-standard type names
			// (e.g. "bool" -> "boolean") that some MCP servers emit.
			var finalTool tool.BaseTool = t
			if tools.NeedsSchemaNormalization(toolInfo.ParamsOneOf) {
				finalTool = tools.NewSchemaNormalizerTool(t)
			}
			base.addToolWithName(toolInfo.Name, finalTool)
		}
	}
	return nil
}

// GetChatModelAgent
func (base *BaseAgent) GetChatModelAgent() (result *adk.ChatModelAgent, err error) {
	// Add error handler middleware to catch tool execution errors and panics
	// This converts errors to ToolMessage so LLM can continue processing
	errorHandler := middleware.NewToolErrorHandler(nil)
	base.conf.ToolsConfig.ToolCallMiddlewares = append(
		base.conf.ToolsConfig.ToolCallMiddlewares,
		errorHandler.ToToolMiddleware(),
	)

	if base.Plugins != nil && base.Plugins.HasToolHooks() {
		base.conf.ToolsConfig.ToolCallMiddlewares = append(
			base.conf.ToolsConfig.ToolCallMiddlewares,
			base.Plugins.NewToolInterceptorMiddleware(),
		)
	}

	base.conf.ToolsConfig.UnknownToolsHandler = func(ctx context.Context, name, input string) (string, error) {
		return fmt.Sprintf("Tool '%s' does not exist. Do not retry or invent tools. Check available tools and retry.", name), nil
	}

	//tools检查中间件
	pcc, err := patchtoolcalls.New(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	base.conf.Handlers = append(base.conf.Handlers, pcc)

	err = base.RegisterLocalFileSystem()
	if err != nil {
		return
	}

	if err = base.addDefaultTools(); err != nil {
		return nil, err
	}

	base.conf.Handlers = append(base.conf.Handlers, base.middlewares...)
	base.addTools(base.toolList...)

	if err = base.loadMCPTools(context.Background()); err != nil {
		return
	}

	if base.conf.Name == "main_agent" {
		for _, t := range base.conf.ToolsConfig.Tools {
			info, e := t.Info(context.Background())
			if e != nil {
				continue
			}
			freedom.Logger().Debugf("main_agent load tool: %s", info.Name)
		}
	}

	result, err = adk.NewChatModelAgent(context.Background(), base.conf)
	return
}

func (base *BaseAgent) addDefaultTools() (err error) {
	if base.sysCfg.WebFetchEnabled {
		wftool, err := tools.NewWebFetch()
		if err != nil {
			return err
		}
		base.addTools(wftool)
	}

	checkFileTool, err := tools.NewCheckFileExists()
	if err != nil {
		return err
	}
	base.addTools(checkFileTool)

	timeTool, err := tools.NewGetCurrentTime()
	if err != nil {
		return err
	}
	base.addTools(timeTool)
	return nil
}

func (base *BaseAgent) RegisterLocalFileSystem() (err error) {
	validateFn := base.validateCommandFunc
	if validateFn == nil {
		validateFn = ValidateCommand
	}

	backend, err := base.sandbox.NewBackend()
	if err != nil {
		return
	}

	shellTimeout := time.Duration(base.sysCfg.ShellTimeoutMinutes) * time.Minute
	if base.conf.Name == systemAgentName {
		shellTimeout = 10 * time.Minute
	}
	shell, err := base.sandbox.NewStreamingShell(&sandbox.ShellConfig{
		ValidateCommand: validateFn,
		Timeout:         shellTimeout,
	})

	backedmid, e := mfilesystem.New(context.Background(), &mfilesystem.MiddlewareConfig{
		Backend:        backend,
		StreamingShell: shell,
	})
	if e != nil {
		err = e
	}
	base.conf.Handlers = append(base.conf.Handlers, backedmid) //用户文件系统和shell

	createDirTool, err := tools.NewCreateDirectory(base.sandbox)
	if err != nil {
		return err
	}
	base.addTools(createDirTool)

	moveFileTool, err := tools.NewMoveFile(base.sandbox)
	if err != nil {
		return err
	}
	base.addTools(moveFileTool)

	deleteFileTool, err := tools.NewDeleteFile(base.sandbox)
	if err != nil {
		return err
	}
	base.conf.ToolsConfig.Tools = append(base.conf.ToolsConfig.Tools, deleteFileTool)

	copyFileTool, err := tools.NewCopyFile(base.sandbox)
	if err != nil {
		return err
	}
	base.conf.ToolsConfig.Tools = append(base.conf.ToolsConfig.Tools, copyFileTool)

	//skills
	var userSkillCfg *middleware.SkillBackendConfig
	userSkillCfg = &middleware.SkillBackendConfig{
		Backend: backend,
		BaseDir: base.sandbox.GetWorkspace() + "/skills",
	}

	var skillBackend skill.Backend
	var skillErr error
	if base.systemSkillProvider != nil {
		sysBackend, e := newSystemSkillBackend(base.systemSkillProvider)
		if e != nil {
			err = e
			return
		}
		skillBackend, skillErr = middleware.NewCombinedSkillBackendWithSystemBackend(context.Background(), userSkillCfg, sysBackend, base.skillAccessor)
	} else {
		skillBackend, skillErr = middleware.NewCombinedSkillBackend(context.Background(), userSkillCfg, nil, base.skillAccessor)
	}
	if skillErr != nil {
		err = skillErr
		return
	}

	//技能中间件加入合并技能
	skillmid, e := skill.NewMiddleware(context.Background(), &skill.Config{
		Backend: skillBackend,
	})
	if e != nil {
		err = e
		return
	}
	base.conf.Handlers = append(base.conf.Handlers, skillmid)
	return
}
