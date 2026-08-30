# GoRaven 插件开发指南

GoRaven 插件系统允许第三方开发者扩展 Agent 行为，无需修改核心代码。遵循"放进去、注册、就能用"的原则。

## 目录结构

```
plugins/                   # 三方扩展目录
├── plugins.go            # 注册入口，main() 调用 RegisterAll()
├── README.md             # 本指南
├── README_EN.md          # 英文指南
├── builtin/
│   └── audit/             # 审计日志示例插件
│       └── audit.go
└── your-extension/        # ★ 在这里添加你的插件
    └── your_plugin.go
```

## 快速入门：创建一个插件

### 第 1 步：创建插件目录和文件

```bash
mkdir -p plugins/my-awesome-plugin
```

### 第 2 步：写插件代码

```go
// plugins/my-awesome-plugin/my_plugin.go
package awesome

import (
    "context"
    "goraven/core/plugin"
)

// Register 注册插件工厂，由 plugins/plugins.go 调用
func Register() {
    plugin.Register(func() plugin.Plugin { return &MyPlugin{} })
}

type MyPlugin struct{}

func (p *MyPlugin) Name() string    { return "my/awesome-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
```

### 第 3 步：实现需要的 Hook 接口

一个 struct 可以同时实现多个 Hook 接口，按需组合：

```go
// === RoundHook: 在每次对话轮次前后执行 ===
func (p *MyPlugin) BeforeRound(ctx *plugin.RoundContext) error {
    // 此时可以注入 system message、修改历史消息
    return nil
}

func (p *MyPlugin) AfterRound(ctx *plugin.RoundContext) error {
    // 此时可以读取回复、推理内容，做审计/统计
    return nil
}

// === ToolHook: 拦截工具执行 ===
func (p *MyPlugin) BeforeTool(ctx context.Context, toolName string, args string) (string, bool) {
    // 返回 skip=true 阻断工具执行
    return args, false
}

func (p *MyPlugin) AfterTool(ctx context.Context, toolName string, args string, result string, execErr error) string {
	// 可修改工具返回值
	return result
}

// === SSEHook: 拦截发给前端的实时事件 ===
func (p *MyPlugin) OnSSEEvent(ctx context.Context, event *plugin.SSEEventData) *plugin.SSEEventData {
    // 返回 nil 过滤掉此事件
    return event
}

// === AgentLifecycleHook: Agent 创建时注入工具/中间件 ===
func (p *MyPlugin) BeforeAgentCreate(ctx *plugin.AgentCreateContext) error {
    // ctx.AddTool(myTool)  — 注入自定义工具
    // ctx.AddMiddleware(myMw) — 注入中间件
    return nil
}
```

### 第 4 步：在 plugins/plugins.go 中激活

在 `plugins/plugins.go` 中 import 你的插件并调用其 `Register()`：

```go
package plugins

import (
    "goraven/plugins/builtin/audit"
    "goraven/plugins/my-awesome-plugin"    // ★ import 你的插件
)

func RegisterAll() {
    audit.Register()
    awesome.Register()                    // ★ 调用注册
}
```

> 所有插件注册收敛在 `plugins/plugins.go` 的 `RegisterAll()` 中，由 `main()` 显式调用。升级核心时不涉及 main.go。
```

### 第 5 步：编译

```bash
go build
```

## 接口参考

### Plugin（基础接口）

所有插件必须实现此接口。

| 方法 | 说明 |
|------|------|
| `Name() string` | 插件唯一标识，建议格式 `"vendor/plugin-name"` |
| `Version() string` | 语义化版本号 |

### RoundHook

每次用户对话轮次的前后触发。覆盖度最高，适合审计、合规、统计等场景。

#### RoundContext 字段

| 字段 | BeforeRound | AfterRound | 说明 |
|------|-------------|------------|------|
| `SessionID` | ✓ | ✓ | 会话 ID |
| `UserID` | ✓ | ✓ | 用户 ID |
| `RoundID` | ✓ | ✓ | 轮次 ID |
| `Query` | ✓ | — | 用户输入 |
| `Messages` | 可读可改 | 只读 | 对话历史 `[]*schema.Message` |
| `Reply` | — | 可读可改 | Agent 回复正文 |
| `ReasoningContent` | — | 只读 | 完整推理/思考链 |
| `Stopped` | — | 只读 | 轮次是否被终止 |

### ToolHook

每次工具执行前后触发，支持参数修改、返回结果修改、短路跳过。

- `BeforeTool` → 可修改参数、可设置 skip=true 阻断执行
- `AfterTool` → 可修改返回结果

### SSEHook

每个 SSE 事件推送前触发，支持内容过滤和替换。返回 `nil` 可完全抑制该事件。

### AgentLifecycleHook

Agent 创建时触发，可注入自定义工具和中间件。实现 `BeforeAgentCreate` 后通过 `ctx.AddTool()` 和 `ctx.AddMiddleware()` 注入。

## 错误处理约定

| Hook | 出错行为 |
|------|----------|
| RoundHook.BeforeRound | 阻断本轮对话，错误文本作为回复 |
| RoundHook.AfterRound | 仅记录日志，不影响回复 |
| ToolHook.BeforeTool | skip=true 阻断，修改后的参数作为工具结果返回 |
| ToolHook.AfterTool | 仅记录日志，原返回值不变 |
| SSEHook.OnSSEEvent | 返回 nil 过滤事件 |
| AgentLifecycleHook | 阻断 Agent 创建 |

## 注意事项

1. 插件通过导出 `Register()` 函数，由 `plugins/plugins.go` 的 `RegisterAll()` 显式调用注册
2. 一个 struct 实现多个接口时，所有接口的 Hook 都会被调用
3. Hook 按注册顺序执行（import 顺序）
4. 插件包应尽量精简，避免引入不必要的第三方依赖
5. 插件执行在主 goroutine 中，避免耗时阻塞

## 示例：审计日志插件

参见 `builtin/audit/audit.go`，展示了 RoundHook + ToolHook + SSEHook 的组合用法。
