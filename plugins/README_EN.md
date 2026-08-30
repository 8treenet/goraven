# GoRaven Plugin Development Guide

The GoRaven plugin system allows third-party developers to extend agent behavior without modifying core code. Drop in, register, and go.

## Directory Structure

```
plugins/                   # Third-party extensions directory
├── plugins.go            # Registration entry, called by main()
├── README.md             # Chinese guide
├── README_EN.md          # This guide
├── builtin/
│   └── audit/             # Reference audit plugin
│       └── audit.go
└── your-extension/        # ★ Add your plugin here
    └── your_plugin.go
```

## Quick Start: Creating a Plugin

### Step 1: Create plugin directory and file

```bash
mkdir -p plugins/my-awesome-plugin
```

### Step 2: Write the plugin code

```go
// plugins/my-awesome-plugin/my_plugin.go
package awesome

import (
    "context"
    "goraven/core/plugin"
)

// Register registers this plugin factory; called by plugins/plugins.go.
func Register() {
    plugin.Register(func() plugin.Plugin { return &MyPlugin{} })
}

type MyPlugin struct{}

func (p *MyPlugin) Name() string    { return "my/awesome-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
```

### Step 3: Implement the hook interfaces you need

A single struct can implement multiple hook interfaces—mix and match as needed:

```go
// === RoundHook: runs before and after each conversation round ===
func (p *MyPlugin) BeforeRound(ctx *plugin.RoundContext) error {
    // Inject system messages, modify history
    return nil
}

func (p *MyPlugin) AfterRound(ctx *plugin.RoundContext) error {
    // Read reply and reasoning, do auditing/analytics
    return nil
}

// === ToolHook: intercept tool execution ===
func (p *MyPlugin) BeforeTool(ctx context.Context, toolName string, args string) (string, bool) {
    // Return skip=true to block tool execution
    return args, false
}

func (p *MyPlugin) AfterTool(ctx context.Context, toolName string, args string, result string, execErr error) string {
	// Modify the tool result
	return result
}

// === SSEHook: intercept real-time events sent to the frontend ===
func (p *MyPlugin) OnSSEEvent(ctx context.Context, event *plugin.SSEEventData) *plugin.SSEEventData {
    // Return nil to suppress this event
    return event
}

// === AgentLifecycleHook: inject tools/middleware during agent creation ===
func (p *MyPlugin) BeforeAgentCreate(ctx *plugin.AgentCreateContext) error {
    // ctx.AddTool(myTool)      — inject a custom tool
    // ctx.AddMiddleware(myMw)  — inject a middleware
    return nil
}
```

### Step 4: Activate in plugins/plugins.go

Import your plugin in `plugins/plugins.go` and call its `Register()`:

```go
package plugins

import (
    "goraven/plugins/builtin/audit"
    "goraven/plugins/my-awesome-plugin"    // ★ import your plugin
)

func RegisterAll() {
    audit.Register()
    awesome.Register()                    // ★ call registration
}
```

> All plugin registrations converge in `plugins/plugins.go`'s `RegisterAll()`, called explicitly by `main()`. Core upgrades don't touch main.go.
```

### Step 5: Build

```bash
go build
```

## Interface Reference

### Plugin (base interface)

All plugins must implement this interface.

| Method | Description |
|--------|-------------|
| `Name() string` | Unique identifier, recommended format: `"vendor/plugin-name"` |
| `Version() string` | Semantic version string |

### RoundHook

Triggered before and after each user conversation round. Highest coverage, suitable for auditing, compliance, and analytics.

#### RoundContext Fields

| Field | BeforeRound | AfterRound | Description |
|-------|-------------|------------|-------------|
| `SessionID` | ✓ | ✓ | Session ID |
| `UserID` | ✓ | ✓ | User ID |
| `RoundID` | ✓ | ✓ | Round ID |
| `Query` | ✓ | — | User input |
| `Messages` | read-write | read-only | Conversation history `[]*schema.Message` |
| `Reply` | — | read-write | Agent reply content |
| `ReasoningContent` | — | read-only | Full reasoning/thinking chain |
| `Stopped` | — | read-only | Whether round was terminated |

### ToolHook

Triggered before and after each tool execution. Supports argument modification, result modification, and short-circuit skip.

- `BeforeTool` — modify arguments, set skip=true to block execution
- `AfterTool` — modify the returned result

### SSEHook

Triggered before each SSE event is pushed to the frontend. Supports content filtering and replacement. Return `nil` to suppress an event entirely.

### AgentLifecycleHook

Triggered during agent creation. Inject custom tools and middleware via `ctx.AddTool()` and `ctx.AddMiddleware()`.

## Error Handling Convention

| Hook | Error Behavior |
|------|---------------|
| RoundHook.BeforeRound | Blocks the round; error text becomes reply |
| RoundHook.AfterRound | Logged only; reply unaffected |
| ToolHook.BeforeTool | skip=true blocks execution; modified args returned as tool result |
| ToolHook.AfterTool | Logged only; original result preserved |
| SSEHook.OnSSEEvent | nil return suppresses the event |
| AgentLifecycleHook | Blocks agent creation |

## Notes

1. Plugins export a `Register()` function, called explicitly from `plugins/plugins.go`'s `RegisterAll()`
2. When a struct implements multiple interfaces, hooks from all interfaces are invoked.
3. Hooks run in registration order (determined by import order).
4. Keep plugin packages minimal. Avoid unnecessary third-party dependencies.
5. Plugins execute on the main goroutine. Avoid blocking operations.

## Reference: Audit Log Plugin

See `builtin/audit/audit.go` for a complete example combining RoundHook + ToolHook + SSEHook.
