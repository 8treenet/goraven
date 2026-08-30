package agent

import (
	"context"
	"errors"
	"goraven/core/iface"
	"goraven/core/middleware/pruning"
	"goraven/core/plugin"
	"goraven/core/sandbox"
	"goraven/core/tools"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/plantask"
	einotool "github.com/cloudwego/eino/components/tool"
)

func init() {
	adk.SetLanguage(adk.LanguageChinese)
}

const mainAgentName = "main_agent"

func NewMainAgent(param AgentParam) (*MainAgent, error) {
	box, boxerr := sandbox.NewSandbox(param.UserName)
	if boxerr != nil {
		return nil, boxerr
	}
	param.userSpace = box.GetWorkspace()
	err := param.Validate()
	if err != nil {
		return nil, err
	}

	// 团队项目：仅放行当前项目子目录，而非整个团队项目根目录，避免跨项目读写
	if param.ProjectWorkspace != "" && param.Project != "" {
		box.SetExtraWorkspace(filepath.Join(param.ProjectWorkspace, param.Project))
	}

	//fmt.Println(getMainInstructionPrompt(param))
	baseAgent := NewBaseAgent(mainAgentName, getMainInstructionPrompt(param), getMainAgentDescription(), param.ChatModel, param.SysCfg, box)

	baseAgent.Plugins = plugin.CreatePlugins()
	baseAgent.Plugins.FireAgentBeforeCreate(&plugin.AgentCreateContext{
		SessionID:     param.SessionId(),
		UserID:        param.UserId(),
		AddTool:       func(t any) { baseAgent.AddTool(t.(einotool.BaseTool)) },
		AddMiddleware: func(m any) { baseAgent.AddChatModelAgentMiddleware(m.(adk.ChatModelAgentMiddleware)) },
	})

	return &MainAgent{
		BaseAgent:       *baseAgent,
		msgRepo:         param.MsgRepo,
		flashModel:      param.ChatModel,
		param:           param,
		box:             box,
		enableStreaming: true,
	}, nil
}

type MainAgent struct {
	BaseAgent
	msgRepo         iface.MessageRepo
	compress        *Compress
	flashModel      iface.BaseChatModel
	visualModel     iface.BaseChatModel
	param           AgentParam
	box             sandbox.Sandbox
	runner          *MainRunner      // 保存 runner 引用，用于重试回调发送 SSE 事件
	planTaskBackend *inMemoryBackend // plan task 中间件的内存 backend，供 SSE 层查询任务 subject
	enableStreaming bool             // Runner 是否启用流式输出，默认 true；静默后台任务应关闭
}

// SetEnableStreaming 设置 Runner 是否启用流式输出。
// 静默后台任务（自动化任务等）无 SSE 消费方，应关闭流式以走非流式消息通道。
func (main *MainAgent) SetEnableStreaming(enable bool) {
	main.enableStreaming = enable
}

func (main *MainAgent) SetFlashModel(model iface.BaseChatModel) {
	main.flashModel = model
}

func (main *MainAgent) SetVisualModel(model iface.BaseChatModel) {
	main.visualModel = model
}

// modelRetryConfig 返回模型重试配置，IsRetryAble 内回调 onModelRetry 通知前端
func (main *MainAgent) modelRetryConfig() *adk.ModelRetryConfig {
	var retryErr error
	var retryErrMutex sync.Mutex

	return &adk.ModelRetryConfig{
		MaxRetries: main.sysCfg.ModelMaxRetries,
		BackoffFunc: func(ctx context.Context, attempt int) time.Duration {
			retryErrMutex.Lock()
			err := retryErr
			retryErr = nil
			retryErrMutex.Unlock()
			main.onModelRetry(main.sysCfg.ModelMaxRetries, attempt, err)
			if err != nil {
				errStr := strings.ToLower(err.Error())
				if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "too many requests") {
					time.Sleep(time.Duration(main.sysCfg.ModelRateLimitWaitSec) * time.Second)
				}
			}
			return time.Duration(attempt) * time.Duration(main.sysCfg.ModelBackoffBaseSec) * time.Second
		},
		IsRetryAble: func(ctx context.Context, err error) bool {
			if err == nil {
				return false
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return false
			}
			retryErrMutex.Lock()
			retryErr = err
			retryErrMutex.Unlock()
			return true
		},
	}
}

// onModelRetry 模型重试回调，通过 SSE 事件通知前端重试状态
func (main *MainAgent) onModelRetry(maxRetries, attempt int, err error) {
	if main.runner == nil {
		return
	}
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	main.runner.sendSSEEvent(&SSEEvent{
		Type: SSEEventTypeRetry,
		Retry: &SSERetryInfo{
			MaxRetries: maxRetries,
			Attempt:    attempt,
			Error:      errStr,
		},
	})
}

func (main *MainAgent) NewRunner(ctx context.Context) (runner *MainRunner, e error) {
	if main.visualModel != nil {
		visualTool, err := tools.NewVisualUnderstand(main.param.UserId(), main.visualModel, main.box, main.param.DailyStatsRepo)
		if err != nil {
			return nil, err
		}
		main.AddTool(visualTool)
	}

	// 创建自动化任务工具：注入持久化仓储与会话暂存上下文（MCP/技能/项目/模型/角色）
	if main.param.AutomationTaskRepo != nil && main.param.Session != nil {
		staging := tools.AutomationStaging{
			McpIds:          main.param.TaskMcpIds,
			SkillIds:        main.param.TaskSkillIds,
			Project:         main.param.Session.Project,
			SharedProjectId: main.param.Session.SharedProjectId,
			AIModelId:       main.param.Session.AIModelId,
			PersonaId:       main.param.Session.PersonaId,
		}
		autoTool, err := tools.NewAutomationTaskCreator(main.param.UserId(), staging, main.param.AutomationTaskRepo)
		if err == nil {
			main.AddTool(autoTool)
		}

		// 查询自动化任务工具：仅返回启用中的任务摘要
		listTool, lerr := tools.NewAutomationTaskLister(main.param.UserId(), main.param.AutomationTaskRepo)
		if lerr == nil {
			main.AddTool(listTool)
		}

		// 任务详情工具：返回含需求描述的完整字段，供修改前回显
		getTool, gerr := tools.NewAutomationTaskGetter(main.param.UserId(), main.param.AutomationTaskRepo)
		if gerr == nil {
			main.AddTool(getTool)
		}

		// 任务更新工具：全量替换业务字段，计划变更时重算下次执行时间
		updateTool, uerr := tools.NewAutomationTaskUpdater(main.param.UserId(), main.param.AutomationTaskRepo)
		if uerr == nil {
			main.AddTool(updateTool)
		}
	}

	// workspace := config.Get().GetUserSpace(main.param.UserId())
	// main.SetWorkspace(workspace)

	subChatAgent, err := main.getSubAgent()
	if err != nil {
		e = err
	}
	subAgentTool := adk.NewAgentTool(ctx, subChatAgent)
	main.AddTool(subAgentTool)
	// Enable internal events streaming so sub-agent's reasoning can be sent to frontend
	main.SetEmitInternalEvents(true)

	// main.SetSKillBaseDir(config.Get().GetUserSpace(main.param.UserId()) + "/skills")
	if main.param.SystemSkillProvider != nil {
		main.SetSystemSkillProvider(main.param.SystemSkillProvider)
	}

	//tools剪枝工具
	pmid, err := pruning.New(main.param.SysCfg)
	if err != nil {
		return nil, err
	}
	main.AddChatModelAgentMiddleware(pmid)

	planTaskMid, err := main.getPlanTask(ctx)
	if err != nil {
		return nil, err
	}
	main.AddChatModelAgentMiddleware(planTaskMid)

	main.compress = NewCompress(main.flashModel, main.msgRepo, main.param.DailyStatsRepo, main.param.SessionId(), main.param.UserId(), main.param.SysCfg)

	// 在调用 GetChatModelAgent 之前设置模型重试配置
	main.SetModelRetryConfig(main.modelRetryConfig())

	agent, err := main.GetChatModelAgent()
	if err != nil {
		return nil, err
	}

	runner = newMainRunner(main, agent, main.enableStreaming)
	main.runner = runner
	return
}

// planTaskBaseDir 是 plan task 中间件持久化任务文件使用的目录，
// 与 inMemoryBackend 中的 key 拼装保持一致，lookupTaskSubject 也按此路径读取。
const planTaskBaseDir = "/tmp/tasks"

func (main *MainAgent) getPlanTask(ctx context.Context) (adk.ChatModelAgentMiddleware, error) {
	backend := newInMemoryBackend()
	main.planTaskBackend = backend
	return plantask.New(ctx, &plantask.Config{
		Backend: backend,
		BaseDir: planTaskBaseDir,
	})
}

func (main *MainAgent) getSubAgent() (result *adk.ChatModelAgent, err error) {
	subAgentModel := main.chatModel
	if main.flashModel != nil {
		subAgentModel = main.flashModel
	}
	subAgent := NewSubAgent(main.param, subAgentModel, main.box)
	if main.param.SystemSkillProvider != nil {
		subAgent.SetSystemSkillProvider(main.param.SystemSkillProvider)
	}

	for _, v := range main.mcpList {
		subAgent.AddMCP(v)
	}
	for _, v := range main.toolList {
		subAgent.AddTool(v)
	}
	for _, v := range main.middlewares {
		subAgent.AddChatModelAgentMiddleware(v)
	}

	pmid, err := pruning.New(main.param.SysCfg)
	if err != nil {
		return nil, err
	}
	subAgent.AddChatModelAgentMiddleware(pmid)

	// 设置与 MainAgent 相同的模型重试回调
	subAgent.SetModelRetryConfig(main.modelRetryConfig())

	result, err = subAgent.GetChatModelAgent()
	return
}
