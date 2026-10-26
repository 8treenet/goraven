package provider

import (
	"errors"

	"goraven/core/iface"
)

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

type Provider interface {
	// ProviderInfo 供应商信息
	Name() string // 供应商标识，如 "deepseek", "glm"

	// Models 返回支持的模型列表
	Models() ([]ModelInfo, error)

	// CreateModel 根据模型名称创建 BaseChatModel 实例
	// modelName: 模型名称
	// temperature: 温度参数，0 表示使用默认值
	CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error)

	//设置代理
	SetProxy(url string) error
}

// ProviderConfig 用于创建 Provider 实例的配置
type ProviderConfig struct {
	APIKey      string // API 密钥
	BaseURL     string // 基础 URL (用于 Ollama, OpenAI 兼容等)
	ExtraFields string // 额外字段
}

// GetProviderByName 根据供应商名称获取 Provider 实例
// 支持的名称: deepseek, bailian (qwen), glm, minimax, openrouter, ollama, openai, claude, gemini, volcano, openai_compatible, claude_compatible
func GetProviderByName(name string, cfg ProviderConfig) (Provider, error) {
	switch name {
	case DeepseekProviderName:
		return NewDeepSeekProvider(cfg.APIKey, cfg.BaseURL), nil
	case QwenProviderName:
		return NewQwenProvider(cfg.APIKey, cfg.BaseURL), nil
	case GLMProviderName:
		return NewGLMProvider(cfg.APIKey, cfg.BaseURL), nil
	case MiniMaxProviderName:
		return NewMiniMaxProvider(cfg.APIKey, cfg.BaseURL), nil
	case OpenrouterProviderName:
		return NewOpenrouterProvider(cfg.APIKey), nil
	case OllamaProviderName:
		return NewOllamaProvider(cfg.BaseURL), nil
	case OpenAIProviderName:
		return NewOpenAIProvider(cfg.APIKey), nil
	case ClaudeProviderName:
		return NewClaudeProvider(cfg.APIKey), nil
	case GeminiProviderName:
		return NewGeminiProvider(cfg.APIKey), nil
	case VolcanoProviderName:
		return NewVolcanoProvider(cfg.APIKey, cfg.BaseURL), nil
	case OpenAICompatibleProviderName:
		return NewOpenAICompatibleProvider(name, cfg.APIKey, cfg.BaseURL, cfg.ExtraFields), nil
	case ClaudeCompatibleProviderName:
		return NewClaudeCompatibleProvider(name, cfg.APIKey, cfg.BaseURL), nil
	default:
		return nil, errors.New("unknown provider name: " + name)
	}
}
