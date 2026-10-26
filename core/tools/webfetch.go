package tools

import (
	"context"
	"fmt"
	"time"

	"goraven/config"
	"goraven/util"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

func init() {
	webFetchClient = requests.NewHTTPClient(time.Duration(config.Get().Tools.HTTPTimeoutSeconds)*time.Second, 5*time.Second)
}

var (
	webFetchClient requests.Client
)

type WebFetchRequest struct {
	URL string `json:"url" jsonschema:"description=The web page URL to fetch, must be a complete HTTP or HTTPS URL"`
}

type WebFetchResponse struct {
	Content string `json:"content" jsonschema:"description=The web page content converted to Markdown format"`
}

type WebFetch struct {
	Name string
	Desc string
}

const (
	WebFetchToolDesc = `Fetches a web page and converts it to Markdown format. Note: This is a simple HTTP GET request without JavaScript rendering capability, so it cannot retrieve content that requires dynamic loading via JavaScript.`

	WebFetchToolDescChinese = `获取网页内容并转换为Markdown格式。注意：这是一个简单的HTTP GET请求，不具备浏览器JS渲染能力，无法获取需要JavaScript动态加载的内容。`
)

func NewWebFetch() (tool.InvokableTool, error) {
	desc := WebFetchToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = WebFetchToolDescChinese
	}

	t := &WebFetch{
		Name: "goraven_web_fetch",
		Desc: desc,
	}

	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

func (w *WebFetch) Invoke(ctx context.Context, req *WebFetchRequest) (*WebFetchResponse, error) {
	if req.URL == "" {
		return nil, fmt.Errorf("URL不能为空")
	}

	html, resp := requests.NewHTTPRequest(req.URL).Get().SetClient(webFetchClient).ToString()
	if resp != nil && resp.Error != nil {
		return nil, fmt.Errorf("请求网页失败: %w", resp.Error)
	}

	markdown, err := util.ConvertMarkDown(html)
	if err != nil {
		return nil, fmt.Errorf("转换Markdown失败: %w", err)
	}

	return &WebFetchResponse{
		Content: markdown,
	}, nil
}
