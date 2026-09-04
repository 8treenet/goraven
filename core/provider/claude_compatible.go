package provider

import (
	"context"
	"goraven/core/iface"
	"goraven/util"
	"net/url"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cloudwego/eino-ext/components/model/claude"
)

const ClaudeCompatibleProviderName = "claude_compatible"

func NewClaudeCompatibleProvider(providerName, apiKey string, baseURL string) *ClaudeCompatibleProvider {
	return &ClaudeCompatibleProvider{
		APIKey:       apiKey,
		BaseURL:      baseURL,
		providerName: providerName,
	}
}

type ClaudeCompatibleProvider struct {
	APIKey       string
	BaseURL      string
	providerName string
	proxyURL     string
}

func (provider *ClaudeCompatibleProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &claude.Config{
		APIKey:     provider.APIKey,
		Model:      modelName,
		BaseURL:    &provider.BaseURL,
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

	return &ClaudeCompatibleModel{
		ChatModel:    *chatModel,
		modelName:    modelName,
		providerName: provider.providerName,
		contextLen:   contextLen,
	}, nil
}

func (provider *ClaudeCompatibleProvider) Name() string {
	return provider.providerName
}

func (provider *ClaudeCompatibleProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []struct {
			ID          string `json:"id"`
			Type        string `json:"type"`
			DisplayName string `json:"display_name"`
			CreatedAt   string `json:"created_at"`
		} `json:"data"`
	}
	err := requests.NewHTTPRequest(provider.BaseURL+"/models?limit=1000").Get().
		SetHeaderValue("x-api-key", provider.APIKey).
		SetHeaderValue("anthropic-version", "2023-06-01").ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	result := make([]ModelInfo, len(body.Data))
	for i, m := range body.Data {
		result[i] = ModelInfo{ID: m.ID, Object: m.Type, OwnedBy: "anthropic"}
	}
	return result, nil
}

func (provider *ClaudeCompatibleProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

type ClaudeCompatibleModel struct {
	claude.ChatModel
	iface.ConversationHeaderKeyHolder
	modelName    string
	providerName string
	contextLen   int
}

func (m *ClaudeCompatibleModel) ModelName() string {
	return m.modelName
}

func (m *ClaudeCompatibleModel) Provider() string {
	return m.providerName
}

func (m *ClaudeCompatibleModel) ContextLength() int {
	return m.contextLen
}

func (m *ClaudeCompatibleModel) Format() iface.APIFormat {
	return iface.APIFormatClaude
}
