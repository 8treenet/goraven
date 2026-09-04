package provider

import (
	"context"
	"goraven/core/iface"
	"goraven/util"
	"net/url"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cloudwego/eino-ext/components/model/claude"
)

const MiniMaxProviderName = "minimax"

func NewMiniMaxProvider(apiKey string, baseURL string) *MiniMaxProvider {
	return &MiniMaxProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

type MiniMaxProvider struct {
	APIKey   string
	BaseURL  string
	proxyURL string
}

// CreateModel
func (provider *MiniMaxProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.minimaxi.com/anthropic"
	}
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &claude.Config{
		APIKey:     provider.APIKey,
		Model:      modelName,
		BaseURL:    &baseURL,
		HTTPClient: httpclient,
	}
	if reasoning {
		config.MaxTokens = 64000
		config.ThinkingConfig = &anthropic.ThinkingConfigParamUnion{
			OfEnabled: &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: 32000,
			},
		}
	}

	chatModel, err := claude.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &MiniMaxModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *MiniMaxProvider) Name() string {
	return MiniMaxProviderName
}

func (provider *MiniMaxProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

// ClaudeModelResponse represents the response from Anthropic models API
type ClaudeModelResponse struct {
	Data []struct {
		Type            string `json:"type"`
		ID              string `json:"id"`
		DisplayName     string `json:"display_name"`
		CreatedAt       string `json:"created_at"`
		MaxInputTokens  int    `json:"max_input_tokens"`
		MaxOutputTokens int    `json:"max_output_tokens"`
	} `json:"data"`
	HasMore bool   `json:"has_more"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
}

// Models returns models from Anthropic API
func (provider *MiniMaxProvider) Models() ([]ModelInfo, error) {
	result := []ModelInfo{}
	result = append(result, ModelInfo{
		ID: "MiniMax-M3",
	})
	result = append(result, ModelInfo{
		ID: "MiniMax-M2.1-highspeed	",
	})
	result = append(result, ModelInfo{
		ID: "MiniMax-M2.7",
	})
	result = append(result, ModelInfo{
		ID: "MiniMax-M2.7-highspeed",
	})
	result = append(result, ModelInfo{
		ID: "MiniMax-M2.5",
	})
	result = append(result, ModelInfo{
		ID: "MiniMax-M2.5-highspeed",
	})
	return result, nil
}

type MiniMaxModel struct {
	claude.ChatModel
	iface.ConversationHeaderKeyHolder
	modelName  string
	contextLen int
}

func (m *MiniMaxModel) ModelName() string {
	return m.modelName
}

func (m *MiniMaxModel) Provider() string {
	return MiniMaxProviderName
}

func (m *MiniMaxModel) ContextLength() int {
	return m.contextLen
}

func (m *MiniMaxModel) Format() iface.APIFormat {
	return iface.APIFormatClaude
}
