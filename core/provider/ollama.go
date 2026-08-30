package provider

import (
	"context"
	"fmt"
	"net/url"
	"goraven/core/iface"
	"goraven/util"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino-ext/components/model/ollama"
)

const OllamaProviderName = "ollama"

func NewOllamaProvider(baseURL string) *OllamaProvider {
	url := baseURL
	if url == "" {
		url = "http://localhost:11434"
	}
	return &OllamaProvider{
		BaseURL: url,
	}
}

type OllamaProvider struct {
	BaseURL  string
	proxyURL string
}

// CreateModel
func (provider *OllamaProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &ollama.ChatModelConfig{
		BaseURL:    provider.BaseURL,
		Model:      modelName,
		HTTPClient: httpclient,
	}

	if reasoning {
		config.Thinking = &ollama.ThinkValue{
			Value: true,
		}
	}

	chatModel, err := ollama.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &OllamaModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *OllamaProvider) Name() string {
	return OllamaProviderName
}

func (provider *OllamaProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

// OllamaModelInfo represents a model from Ollama API
type OllamaModelInfo struct {
	Name string `json:"name"`
}

// Models returns models from Ollama API
func (provider *OllamaProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Models []OllamaModelInfo `json:"models"`
	}
	url := fmt.Sprintf("%s/api/tags", provider.BaseURL)
	err := requests.NewHTTPRequest(url).Get().ToJSON(&body).Error
	if err != nil {
		return nil, err
	}

	result := make([]ModelInfo, len(body.Models))
	for i, m := range body.Models {
		result[i] = ModelInfo{
			ID:      m.Name,
			Object:  "model",
			OwnedBy: "ollama",
		}
	}
	return result, nil
}

type OllamaModel struct {
	ollama.ChatModel
	modelName  string
	contextLen int
}

func (m *OllamaModel) ModelName() string {
	return m.modelName
}

func (m *OllamaModel) Provider() string {
	return OllamaProviderName
}

func (m *OllamaModel) ContextLength() int {
	return m.contextLen
}

func (m *OllamaModel) Format() iface.APIFormat {
	return iface.APIFormatOllama
}
