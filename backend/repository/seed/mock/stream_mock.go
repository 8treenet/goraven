package mock

import (
	"strings"
	"time"

	"goraven/core/agent"
)

// StreamMockChannelBuffer 模拟 SSE 事件通道的缓冲大小，
// 与 core/agent.MainRunner.Query 中保持一致，避免长流阻塞。
const StreamMockChannelBuffer = 20000

// StreamMockChunkSize 每个 content 事件包含的字符数。
// 越小越像真实 LLM 流式输出，越大越省 CPU。
const StreamMockChunkSize = 4

// StreamMockChunkDelayMs 连续 content chunk 之间的间隔（毫秒）。
// 配合 ChunkSize=4 时约为每秒 18 字符，模拟中速流式输出。
const StreamMockChunkDelayMs = 220

// StreamMockStepDelayMs 事件之间的默认停顿（毫秒），模拟 Agent “思考” 间隙。
const StreamMockStepDelayMs = 600

// StreamMockToolMaxSleepMs 单个 tool 事件允许的最大 sleep 时长。
// 需求规定不能超过 30 秒。
const StreamMockToolMaxSleepMs = 30 * 1000

// BuildStreamMock 返回一个会按脚本产生 SSE 事件的 channel。
// goroutine 会在脚本走完后自动 close channel，并标记 session 完成。
//
// sessionId 用于脚本结束时调用 MarkStreamCompleted，保证即使前端提前断开 SSE
// 连接（如切入 background 轮询模式），轮询期间也能看到 status=0。
func BuildStreamMock(sessionId string) <-chan *agent.SSEEvent {
	ch := make(chan *agent.SSEEvent, StreamMockChannelBuffer)
	stopHeartbeat := make(chan struct{})
	go runStreamScript(sessionId, ch, stopHeartbeat)
	go runHeartbeat(ch, stopHeartbeat)
	return ch
}

func runStreamScript(sessionId string, ch chan<- *agent.SSEEvent, stopHeartbeat chan struct{}) {
	defer close(stopHeartbeat)
	defer close(ch)
	defer MarkStreamCompleted(sessionId)

	steps := scriptSteps()
	for _, step := range steps {
		step.emit(ch)
	}
}

// runHeartbeat 模拟 main_runner.startHeartbeat() 的行为：
// 每 30 秒往通道里塞一个 heartbeat（前端俗称 ping）事件。
// 在主脚本结束 / 客户端断开时通过 stop 信号立即退出。
func runHeartbeat(ch chan<- *agent.SSEEvent, stop <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			// channel 可能已经被主脚本关闭，做一次非阻塞尝试
			select {
			case ch <- &agent.SSEEvent{Type: agent.SSEEventTypeHeartbeat}:
			default:
			}
		}
	}
}

type streamScriptStep struct {
	kind     string              // reasoning | content | tool | retry | end
	content  string              // reasoning / content 文本
	tool     *agent.SSEEventTool // kind=tool
	retry    *agent.SSERetryInfo // kind=retry
	preSleep time.Duration       // 事件发送前的等待
	postWork time.Duration       // 事件发送后的等待（tool 事件最长 30s）
}

func (s streamScriptStep) emit(ch chan<- *agent.SSEEvent) {
	if s.preSleep > 0 {
		time.Sleep(s.preSleep)
	}
	switch s.kind {
	case "reasoning":
		ch <- &agent.SSEEvent{Type: agent.SSEEventTypeReasoning, Content: s.content}
	case "content":
		streamContent(ch, s.content)
	case "tool":
		ch <- &agent.SSEEvent{Type: agent.SSEEventTypeTool, Tool: s.tool}
	case "retry":
		ch <- &agent.SSEEvent{Type: agent.SSEEventTypeRetry, Retry: s.retry}
	case "end":
		ch <- &agent.SSEEvent{Type: agent.SSEEventTypeEnd}
	}
	if s.postWork > 0 {
		time.Sleep(s.postWork)
	}
}

func streamContent(ch chan<- *agent.SSEEvent, text string) {
	runes := []rune(text)
	for i := 0; i < len(runes); i += StreamMockChunkSize {
		end := i + StreamMockChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		ch <- &agent.SSEEvent{Type: agent.SSEEventTypeContent, Content: string(runes[i:end])}
		if end < len(runes) {
			time.Sleep(time.Duration(StreamMockChunkDelayMs) * time.Millisecond)
		}
	}
}

func scriptSteps() []streamScriptStep {
	const step = StreamMockStepDelayMs * time.Millisecond

	// 工具事件 sleep 时长（不能超过 30s）
	const (
		toolListSleep = 25 * time.Second
		toolReadSleep = 18 * time.Second
		toolGrepSleep = 22 * time.Second
		toolEditSleep = 15 * time.Second
		toolBashSleep = 12 * time.Second
	)

	// 内容块：每块约 1500~1800 字符，
	// 5 字符/chunk × 180ms ≈ 54~65 秒/块；4 块共约 230 秒。
	content1 := strings.Join([]string{
		"经过初步扫描，该项目采用了经典的 Go 工程布局：",
		"顶层包含 cmd/、internal/、pkg/ 三个核心目录，外加 configs/、scripts/、docs/、frontend/、plugins/ 等支持目录。",
		"main.go 位于 cmd/goraven 下，作为单一可执行入口；",
		"internal/ 下按业务域拆分 controller、service、repository 三层，",
		"依赖方向严格自上而下，没有出现跨层反向引用，goimports 也是按这个顺序分组的。",
		"",
		"接下来我会按 ①入口与路由 ②配置加载 ③服务编排 ④数据访问 ⑤Agent 核心 这五条主线，",
		"逐文件审查职责、错误处理、并发安全、测试覆盖度，并对照 freedom 框架的推荐用法校验 DI 注入。",
		"",
		"需要重点关注的几个点：",
		"  · agent/ 目录与 backend/ 是否存在循环依赖（import graph 上是否有反向边）",
		"  · repository 层的 GORM 调用是否在事务边界外执行、是否漏掉 defer rollback",
		"  · freedom 框架的 DI 注入是否被误用为 service locator，导致测试时无法 mock",
		"  · SSE 长连接的取消与心跳实现是否完备，断网/拔线时 goroutine 能否及时退出",
		"  · 配置热更新的写入是否加锁，避免与读协程竞争出现半新半旧状态",
		"  · GORM 的 soft delete 标志位是否在所有读路径上都被正确过滤",
		"  · 日志中是否泄漏了用户的 API Key / Token 明文",
		"",
		"下面我先把 main.go 拉出来通读一遍，再去看 freedom 框架的 BeforeActivation 钩子怎么注册路由。",
		"如果发现 freedom 用了非主流的路由分组方式，我会单独列一节讲它的代价。",
	}, "\n")

	content2 := strings.Join([]string{
		"在错误处理方面，我注意到 service 层统一通过 errs 包下的预定义错误返回，",
		"controller 层通过 infra.JSONResponse 包装，HTTP 状态码由中间件统一映射；",
		"这是一种比较干净的做法，避免了错误信息泄漏到前端，也方便集中维护错误码表。",
		"前端只需要看 code 字段就能直接展示本地化文案，不需要解析 message。",
		"",
		"但仍然有 3 个潜在问题需要你关注：",
		"  1. 多个 service 在底层错误未被包装时直接向上抛，",
		"     freedom 框架会把这个原始 error 写进响应体，",
		"     等于把数据库错误原文（包含 SQL 片段）暴露给了前端，",
		"     XSS 风险不高但信息泄露风险存在。",
		"  2. 部分 defer cleanup 缺少错误判断，",
		"     文件句柄、Redis 连接、数据库事务在 panic 路径上可能泄漏，",
		"     长期运行下会让连接池被打满。",
		"     建议所有 defer 都接一个命名返回值并显式判断。",
		"  3. agent.MainRunner 的 cancelCtx 取消后，",
		"     sseChan 关闭时序与 HTTP 响应 Flush 的竞争，",
		"     当前实现没有显式同步，理论上可能在客户端多收到一个 end 事件。",
		"     需要补一个并发回归测试。",
		"",
		"建议把 service 入口加一道 fmt.Errorf(\"op %s: %w\", name, err) 的统一包装层，",
		"并在 controller 入口处 errors.Is 判断 + 统一 log.WithError 记录。",
		"这样既保留了错误链，又不会让原始错误原文暴露到响应里，",
		"也对未来的 Sentry / OpenTelemetry 接入更友好。",
	}, "\n")

	content3 := strings.Join([]string{
		"并发安全方面，agent.MainRunner 使用了 sync.Mutex 保护 fetchStatus / stop 标志位，",
		"这个写法本身没问题，但 sendSSEEvent 内部又调用了 dispatchSSE，",
		"而 dispatchSSE 会触发 plugin.FireSSEHook — 该 hook 是阻塞调用，",
		"如果某个 plugin hook 实现里有慢 IO",
		"（比如远程审计上报、数据库写入、Prometheus 上报），",
		"整个 SSE 通道都会被拖慢，客户端会看到明显的卡顿。",
		"",
		"建议：",
		"  · 给 FireSSEHook 加上 200ms 的超时控制，超时后降级为 noop 并 log warn；",
		"  · 或者把 plugin 调用移到独立 goroutine 中，把结果通过 channel 回收，",
		"    避免反向影响主流程的实时性；",
		"  · 心跳事件 (SSEEventTypeHeartbeat) 不要经过 plugin 钩子，保持通道纯净；",
		"  · 锁粒度考虑拆成两个：status 锁 + sseChan 锁，",
		"    减少 plugin 钩子执行时锁竞争的概率。",
		"",
		"测试覆盖方面，我看到 backend/service 下有 chat_test.go、session_test.go，",
		"覆盖了主要 service 的正常 / 异常分支。",
		"但 backend/controller/test 目录下的接口测试",
		"（chat_test.go、user_test.go、admin_*.go 等）",
		"没有出现在 CI 流水线中（看 .github/workflows 没有触发该目录的命令）。",
		"建议把它们纳入 PR 必跑集合，",
		"特别是 persona 和 skill 的接口测试，",
		"它们覆盖了复杂的 JSON 嵌套结构，回归风险最高。",
		"",
		"另外，agent/ 目录的测试目前是空缺的，",
		"整个核心循环没有自动化测试覆盖，",
		"这是项目里最大的盲区之一。",
		"建议把 MainRunner.Query 的关键分支（取消 / 超时 / 工具失败）",
		"用 fake chat model 跑一轮 e2e 级别的回归。",
	}, "\n")

	content4 := strings.Join([]string{
		"总结一下：",
		"  · 整体架构清晰，分层合理，命名规范统一，目录边界与导入方向都控制得很干净。",
		"  · 主要风险集中在三处：错误处理包装、plugin 阻塞、SSE 取消时序。",
		"  · 测试方面建议补齐 controller/test 的 CI 接入，",
		"    并给 plugin.SSEHook 加上超时控制；",
		"    agent/ 目录的测试缺口是当前最值得投入人力的地方。",
		"",
		"可以按上面的优先级分两个迭代处理：",
		"  · 第一迭代（紧急）：补错误包装 + plugin 超时 + SSE 取消时序回归测试，",
		"    目标是把生产事故面缩到最小。",
		"  · 第二迭代（长期）：接入 CI 接口测试 + agent/ 核心循环单测 + 文档更新，",
		"    目标是把 PR 合入门槛抬到一个稳定水位。",
		"",
		"需要我直接生成对应的补丁吗？",
		"我可以按第一迭代的优先级给你出一版 PR，",
		"或者先单独把 SSE 取消时序那段 race 修掉，",
		"附一个 reproducer 测试，方便你 review。",
		"如果倾向后者，给我个信号我先开 PR。",
	}, "\n")

	return []streamScriptStep{
		// 0. 用户视角开场白（reasoning）
		{
			kind:     "reasoning",
			content:  "用户希望我对一个 Go 项目做一轮代码审查。我会先扫目录结构，再读关键入口文件，然后按层逐个分析并发、错误处理、测试覆盖度，最后给一份按优先级排序的修复建议。",
			preSleep: step,
		},
		// 1. 工具：列出目录
		{
			kind:     "tool",
			tool:     toolDef("ls", "📁", "文件系统", "正在扫描项目根目录"),
			preSleep: step,
			postWork: toolListSleep,
		},
		// 2. 工具：读取入口
		{
			kind:     "tool",
			tool:     toolDef("read_file", "📄", "文件系统", "正在读取 cmd/goraven/main.go"),
			preSleep: step,
			postWork: toolReadSleep,
		},
		// 3. 第一段分析（≈ 55s）
		{
			kind:     "content",
			content:  content1,
			preSleep: step,
		},
		// 4. 工具：搜索错误处理模式
		{
			kind:     "tool",
			tool:     toolDef("grep", "🔎", "文件系统", "正在搜索错误处理模式"),
			preSleep: step,
			postWork: toolGrepSleep,
		},
		// 5. 第二段分析（≈ 65s）
		{
			kind:     "content",
			content:  content2,
			preSleep: step,
		},
		// 6. 工具：标记可疑代码段
		{
			kind:     "tool",
			tool:     toolDef("edit_file", "📝", "文件系统", "正在标记可疑代码段"),
			preSleep: step,
			postWork: toolEditSleep,
		},
		// 7. 第三段分析（≈ 70s）
		{
			kind:     "content",
			content:  content3,
			preSleep: step,
		},
		// 8. 模型重试事件
		{
			kind:     "retry",
			retry:    &agent.SSERetryInfo{MaxRetries: 3, Attempt: 1, Error: "upstream connection reset by peer"},
			preSleep: step,
			postWork: 3 * time.Second,
		},
		// 9. 工具：执行命令
		{
			kind:     "tool",
			tool:     toolDef("execute", "⚙️", "文件系统", "正在运行 go vet ./..."),
			preSleep: step,
			postWork: toolBashSleep,
		},
		// 10. 第四段分析（≈ 45s）
		{
			kind:     "content",
			content:  content4,
			preSleep: step,
		},
		// 11. 结束
		{
			kind: "end",
		},
	}
}

func toolDef(name, icon, displayName, action string) *agent.SSEEventTool {
	return &agent.SSEEventTool{
		Name:        name,
		Icon:        icon,
		DisplayName: displayName,
		Action:      action,
	}
}
