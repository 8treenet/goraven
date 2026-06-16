package iface

import "github.com/cloudwego/eino/components/model"

type APIFormat string

const (
	APIFormatOpenAI APIFormat = "openai"
	APIFormatClaude APIFormat = "claude"
	APIFormatOllama APIFormat = "ollama"
	APIFormatGemini APIFormat = "gemini"
)

type BaseChatModel interface {
	model.BaseChatModel
	ModelName() string
	Provider() string
	ContextLength() int
	Format() APIFormat
}
