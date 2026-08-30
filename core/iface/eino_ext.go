package iface

import "github.com/cloudwego/eino/components/model"

// APIFormat 模型的多模态 API 格式类型
type APIFormat string

const (
	APIFormatOpenAI APIFormat = "openai"  // OpenAI 兼容格式：图片(base64/URL)、音频(base64)、视频(URL/base64)
	APIFormatClaude APIFormat = "claude"  // Claude 兼容格式：仅图片(base64/URL)
	APIFormatOllama APIFormat = "ollama"  // Ollama 格式：仅图片(base64)
	APIFormatGemini APIFormat = "gemini"  // Gemini 格式：图片/音频/视频/文件(base64/URL)全支持
)

type BaseChatModel interface {
	model.BaseChatModel
	ModelName() string
	Provider() string
	ContextLength() int //模型上下文长度 单位字节
	Format() APIFormat  //模型的多模态 API 格式
}
