package provider_test

import (
	"goraven/core/provider"
	unit_test "goraven/util/unit"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestDeepSeekProvider_DisplayName(t *testing.T) {
	provider := provider.NewDeepSeekProvider("sk-", "")
	models, _ := provider.Models()
	t.Log(unit_test.JsonLog(models))
}

func TestOpenRouterProvider_DisplayName(t *testing.T) {
	provider := provider.NewOpenrouterProvider("sk-or-v1-")
	models, _ := provider.Models()
	t.Log(unit_test.JsonLog(models))
}

func TestOpenRouterProvider_CreateModel(t *testing.T) {
	provider := provider.NewOpenrouterProvider("sk-or-v1-")
	chatmodel, err := provider.CreateModel("google/gemma-4-31b-it", true, 250000)
	if err != nil {
		panic(err)
	}
	data, err := chatmodel.Generate(t.Context(), []*schema.Message{schema.UserMessage("你是谁")})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}

func TestOpenAiProvider_CreateModel(t *testing.T) {
	provider := provider.NewOpenAICompatibleProvider("kimi", "sk-", "https://api.moonshot.cn/v1", "")
	chatmodel, err := provider.CreateModel("kimi-k2-thinking", true, 250000)
	if err != nil {
		panic(err)
	}
	data, err := chatmodel.Generate(t.Context(), []*schema.Message{schema.UserMessage("你是谁")})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}

func TestQwenProvider_DisplayName(t *testing.T) {
	provider := provider.NewQwenProvider("sk-", "")
	models, _ := provider.Models()
	t.Log(unit_test.JsonLog(models))
}

func TestQwenProvider_CreateModel(t *testing.T) {
	provider := provider.NewQwenProvider("sk-", "")
	chatmodel, err := provider.CreateModel("qwen3.6-plus", true, 250000)
	if err != nil {
		panic(err)
	}
	data, err := chatmodel.Generate(t.Context(), []*schema.Message{schema.UserMessage("你是谁")})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}

func TestGLMProvider_DisplayName(t *testing.T) {
	provider := provider.NewGLMProvider(".O3IWA5vn9LHgb1VV", "")
	models, _ := provider.Models()
	t.Log(unit_test.JsonLog(models))
}

func TestGLMProvider_CreateModel(t *testing.T) {
	provider := provider.NewGLMProvider("", "")
	chatmodel, err := provider.CreateModel("glm-4.5", true, 200000)
	if err != nil {
		panic(err)
	}
	data, err := chatmodel.Generate(t.Context(), []*schema.Message{schema.UserMessage("你是谁")})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}

func TestMiniMaxProvider_DisplayName(t *testing.T) {
	provider := provider.NewMiniMaxProvider("", "")
	models, _ := provider.Models()
	t.Log(unit_test.JsonLog(models))
}

func TestMiniMaxProvider_CreateModel(t *testing.T) {
	provider := provider.NewMiniMaxProvider("", "")
	chatmodel, err := provider.CreateModel("MiniMax-M2.5", true, 200000)
	if err != nil {
		panic(err)
	}
	data, err := chatmodel.Generate(t.Context(), []*schema.Message{schema.UserMessage("你是谁")})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}

func TestClaudeProvider_CreateModel(t *testing.T) {
	provider := provider.NewClaudeCompatibleProvider(
		"minimax",
		"",
		"https://api.minimaxi.com/anthropic",
	)
	chatmodel, err := provider.CreateModel("MiniMax-M2.5", true, 200000)
	if err != nil {
		panic(err)
	}
	data, err := chatmodel.Generate(t.Context(), []*schema.Message{schema.UserMessage("你是谁")})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(data))
}
