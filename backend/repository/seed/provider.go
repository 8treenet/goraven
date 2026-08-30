package seed

import (
	"goraven/config"
	"goraven/core/provider"
)

type ProviderDef struct {
	ID               string
	DisplayNameZh    string
	DisplayNameEn    string
	Icon             string
	DefaultBaseURLZh string
	DefaultBaseURLEn string
	RequireAPIKey    bool
	RequireBaseURL   bool
}

func (def *ProviderDef) CurrentDefaultBaseURL() string {
	if config.Get().GetLanguage() == "en" {
		return def.DefaultBaseURLEn
	}
	return def.DefaultBaseURLZh
}

var ProviderDefs = []ProviderDef{
	{
		ID:               provider.DeepseekProviderName,
		DisplayNameZh:    "DeepSeek",
		DisplayNameEn:    "DeepSeek",
		Icon:             "/logos/deepseek.svg",
		DefaultBaseURLZh: "https://api.deepseek.com",
		DefaultBaseURLEn: "https://api.deepseek.com",
		RequireAPIKey:    true,
		RequireBaseURL:   true,
	},
	{
		ID:               provider.QwenProviderName,
		DisplayNameZh:    "Alibaba",
		DisplayNameEn:    "Alibaba",
		Icon:             "/logos/bailian.svg",
		DefaultBaseURLZh: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		DefaultBaseURLEn: "https://dashscope-us.aliyuncs.com/compatible-mode/v1",
		RequireAPIKey:    true,
		RequireBaseURL:   true,
	},
	{
		ID:               provider.GLMProviderName,
		DisplayNameZh:    "Zhipu",
		DisplayNameEn:    "Zhipu",
		Icon:             "/logos/zhipu.svg",
		DefaultBaseURLZh: "https://open.bigmodel.cn/api/paas/v4",
		DefaultBaseURLEn: "https://api.z.ai/api/paas/v4",
		RequireAPIKey:    true,
		RequireBaseURL:   true,
	},
	{
		ID:               provider.MiniMaxProviderName,
		DisplayNameZh:    "MiniMax",
		DisplayNameEn:    "MiniMax",
		Icon:             "/logos/minimax.svg",
		DefaultBaseURLZh: "https://api.minimaxi.com/anthropic",
		DefaultBaseURLEn: "https://api.minimax.io/anthropic",
		RequireAPIKey:    true,
		RequireBaseURL:   true,
	},
	{
		ID:            provider.OpenrouterProviderName,
		DisplayNameZh: "OpenRouter",
		DisplayNameEn: "OpenRouter",
		Icon:          "/logos/openrouter.svg",
		RequireAPIKey: true},
	{
		ID:             provider.OllamaProviderName,
		DisplayNameZh:  "Ollama",
		DisplayNameEn:  "Ollama",
		Icon:           "/logos/ollama.svg",
		RequireBaseURL: true,
	},
	{
		ID:            provider.OpenAIProviderName,
		DisplayNameZh: "OpenAI",
		DisplayNameEn: "OpenAI",
		Icon:          "/logos/openai.svg",
		RequireAPIKey: true,
	},
	{
		ID:            provider.ClaudeProviderName,
		DisplayNameZh: "Anthropic",
		DisplayNameEn: "Anthropic",
		Icon:          "/logos/claude.svg",
		RequireAPIKey: true,
	},
	{
		ID:            provider.GeminiProviderName,
		DisplayNameZh: "Google",
		DisplayNameEn: "Google",
		Icon:          "/logos/gemini.svg",
		RequireAPIKey: true,
	},
	{
		ID:               provider.VolcanoProviderName,
		DisplayNameZh:    "Volcano",
		DisplayNameEn:    "byteplus",
		Icon:             "/logos/huoshan.png",
		DefaultBaseURLZh: "https://ark.cn-beijing.volces.com/api/v3",
		DefaultBaseURLEn: "https://ark.ap-southeast.bytepluses.com/api/v3",
		RequireAPIKey:    true,
		RequireBaseURL:   true,
	},
	{
		ID:             provider.OpenAICompatibleProviderName,
		DisplayNameZh:  "OpenAI Compatible",
		DisplayNameEn:  "OpenAI Compatible",
		RequireAPIKey:  true,
		RequireBaseURL: true,
	},
	{
		ID:             provider.ClaudeCompatibleProviderName,
		DisplayNameZh:  "Claude Compatible",
		DisplayNameEn:  "Claude Compatible",
		RequireAPIKey:  true,
		RequireBaseURL: true,
	},
}
