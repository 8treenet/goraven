package iface

import "github.com/cloudwego/eino/components/model"

// APIFormat 模型的多模态 API 格式类型
type APIFormat string

const (
	APIFormatOpenAI APIFormat = "openai" // OpenAI 兼容格式：图片(base64/URL)、音频(base64)、视频(URL/base64)
	APIFormatClaude APIFormat = "claude" // Claude 兼容格式：仅图片(base64/URL)
	APIFormatOllama APIFormat = "ollama" // Ollama 格式：仅图片(base64)
	APIFormatGemini APIFormat = "gemini" // Gemini 格式：图片/音频/视频/文件(base64/URL)全支持
)

type BaseChatModel interface {
	model.BaseChatModel
	ModelName() string
	Provider() string
	ContextLength() int                  //模型上下文长度 单位字节
	Format() APIFormat                   //模型的多模态 API 格式
	SetConversationHeaderKey(key string) //设置会话归并 header 名（如 X-Opencode-Session），空则不注入
	GetConversationHeaderKey() string    //读取会话归并 header 名
}

// ConversationHeaderKeyHolder 会话归并 header 名的持有者，BaseChatModel 实现内嵌该结构即可获得 Set/Get 能力
type ConversationHeaderKeyHolder struct {
	conversationHeaderKey string
}

func (h *ConversationHeaderKeyHolder) SetConversationHeaderKey(key string) {
	h.conversationHeaderKey = key
}

func (h *ConversationHeaderKeyHolder) GetConversationHeaderKey() string {
	return h.conversationHeaderKey
}
