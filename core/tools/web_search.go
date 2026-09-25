package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"goraven/config"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"golang.org/x/net/html"
)

const duckDuckGoHTMLSearchURL = "https://html.duckduckgo.com/html/"

const (
	WebSearchToolDesc = `Search the public internet for current information without an API key. Returns titles, URLs, and snippets from DuckDuckGo. Use this before answering questions that depend on recent events, changing facts, or online sources.`
	WebSearchToolDescChinese = `使用无需 API Key 的 DuckDuckGo 搜索公开互联网实时信息，返回标题、URL 和摘要。回答依赖最新事件、会变化的事实或在线来源的问题前应优先使用此工具。`
)

type WebSearchRequest struct {
	Query      string `json:"query" jsonschema:"description=Search query, required"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"description=Maximum number of results from 1 to 10, optional, defaults to 5"`
}

type WebSearchResponse struct {
	Content string `json:"content" jsonschema:"description=Clean numbered search results containing title, URL, and snippet"`
}

type webSearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// WebSearch uses DuckDuckGo's public HTML results page. It is deliberately
// keyless: no paid provider account or secret configuration is required.
type WebSearch struct {
	Name     string
	Desc     string
	client   *http.Client
	endpoint string
}

func NewWebSearch() (tool.InvokableTool, error) {
	desc := WebSearchToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = WebSearchToolDescChinese
	}

	t := &WebSearch{
		Name:     "goraven_web_search",
		Desc:     desc,
		client:   &http.Client{Timeout: time.Duration(config.Get().Tools.HTTPTimeoutSeconds) * time.Second},
		endpoint: duckDuckGoHTMLSearchURL,
	}
	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

func (w *WebSearch) Invoke(ctx context.Context, req *WebSearchRequest) (*WebSearchResponse, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	limit := req.MaxResults
	if limit == 0 {
		limit = 5
	}
	if limit < 1 || limit > 10 {
		return nil, fmt.Errorf("max_results must be between 1 and 10")
	}

	endpoint, err := url.Parse(w.endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid search endpoint: %w", err)
	}
	params := endpoint.Query()
	params.Set("q", query)
	params.Set("kl", "wt-wt") // neutral locale keeps result parsing predictable.
	endpoint.RawQuery = params.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}
	httpRequest.Header.Set("User-Agent", "GoRaven/1.0 (+https://github.com/8treenet/goraven)")
	httpRequest.Header.Set("Accept", "text/html,application/xhtml+xml")

	response, err := w.client.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("DuckDuckGo search request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("DuckDuckGo search returned HTTP %d", response.StatusCode)
	}

	results, err := parseDuckDuckGoHTML(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("parse DuckDuckGo search results: %w", err)
	}
	if len(results) > limit {
		results = results[:limit]
	}
	return &WebSearchResponse{Content: formatWebSearchResults(query, results)}, nil
}

func parseDuckDuckGoHTML(r io.Reader) ([]webSearchResult, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var results []webSearchResult
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "div" && hasClass(node, "result") {
			result := webSearchResult{
				Title:   nodeText(findDescendantByClass(node, "result__a")),
				URL:     searchResultURL(findDescendantByClass(node, "result__a")),
				Snippet: nodeText(findDescendantByClass(node, "result__snippet")),
			}
			if result.Title != "" && result.URL != "" {
				results = append(results, result)
			}
			return // Result nodes are self-contained; avoid nested duplicate matches.
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return results, nil
}

func hasClass(node *html.Node, class string) bool {
	for _, attr := range node.Attr {
		if attr.Key != "class" {
			continue
		}
		for _, value := range strings.Fields(attr.Val) {
			if value == class {
				return true
			}
		}
	}
	return false
}

func findDescendantByClass(node *html.Node, class string) *html.Node {
	if node != nil && hasClass(node, class) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findDescendantByClass(child, class); found != nil {
			return found
		}
	}
	return nil
}

func nodeText(node *html.Node) string {
	if node == nil {
		return ""
	}
	var text strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			text.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(strings.Fields(text.String()), " ")
}

func searchResultURL(anchor *html.Node) string {
	if anchor == nil {
		return ""
	}
	for _, attr := range anchor.Attr {
		if attr.Key != "href" {
			continue
		}
		href := strings.TrimSpace(attr.Val)
		if parsed, err := url.Parse(href); err == nil && parsed.Path == "/l/" {
			if target := parsed.Query().Get("uddg"); target != "" {
				return target
			}
		}
		return href
	}
	return ""
}

func formatWebSearchResults(query string, results []webSearchResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No web search results found for %q.", query)
	}
	var output strings.Builder
	fmt.Fprintf(&output, "Web search results for %q:\n", query)
	for i, result := range results {
		fmt.Fprintf(&output, "\n%d. %s\nURL: %s\n", i+1, result.Title, result.URL)
		if result.Snippet != "" {
			fmt.Fprintf(&output, "Snippet: %s\n", result.Snippet)
		}
	}
	return strings.TrimSpace(output.String())
}
