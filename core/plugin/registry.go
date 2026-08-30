package plugin

import "sync"

type registry struct {
	mu        sync.RWMutex
	factories []func() Plugin
}

var global = &registry{}

// PluginInfo 插件元信息（名称 + 版本）
type PluginInfo struct {
	Name    string
	Version string
}

// Register stores a factory function that creates a new plugin instance.
// Each agent calls the factory to get its own copy.
func Register(factory func() Plugin) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.factories = append(global.factories, factory)
}

// GetAllPluginInfo 返回所有已注册插件的名称和版本（名称+版本列表）
// 用于系统信息页展示已加载插件
func GetAllPluginInfo() []PluginInfo {
	global.mu.RLock()
	defer global.mu.RUnlock()

	var result []PluginInfo
	for _, factory := range global.factories {
		inst := factory()
		result = append(result, PluginInfo{
			Name:    inst.Name(),
			Version: inst.Version(),
		})
	}
	return result
}

// CreatePlugins instantiates a fresh set of plugin instances for a new agent.
// Each call to a factory produces a separate instance, so plugins can hold
// per-agent state without cross-contamination.
func CreatePlugins() *Plugins {
	global.mu.RLock()
	defer global.mu.RUnlock()

	p := &Plugins{}
	for _, factory := range global.factories {
		inst := factory()
		p.plugins = append(p.plugins, inst)
		if h, ok := inst.(AgentLifecycleHook); ok {
			p.lifecycleHooks = append(p.lifecycleHooks, h)
		}
		if h, ok := inst.(RoundHook); ok {
			p.roundHooks = append(p.roundHooks, h)
		}
		if h, ok := inst.(ToolHook); ok {
			p.toolHooks = append(p.toolHooks, h)
		}
		if h, ok := inst.(SSEHook); ok {
			p.sseHooks = append(p.sseHooks, h)
		}
	}
	return p
}
