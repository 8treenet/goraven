package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const duckDuckGoResultsFixture = `<!doctype html><html><body>
<div class="result results_links results_links_deep">
  <h2><a class="result__a" href="https://example.com/one">First &amp; Result</a></h2>
  <a class="result__snippet">A useful <b>first</b> snippet.</a>
</div>
<div class="result results_links results_links_deep">
  <h2><a class="result__a" href="/l/?uddg=https%3A%2F%2Fexample.org%2Ftwo%3Fa%3D1">Second result</a></h2>
  <div class="result__snippet">Second snippet with current information.</div>
</div>
</body></html>`

func TestParseDuckDuckGoHTML(t *testing.T) {
	results, err := parseDuckDuckGoHTML(strings.NewReader(duckDuckGoResultsFixture))
	if err != nil {
		t.Fatalf("parseDuckDuckGoHTML() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("result count = %d, want 2", len(results))
	}
	if got, want := results[0], (webSearchResult{
		Title: "First & Result", URL: "https://example.com/one", Snippet: "A useful first snippet.",
	}); got != want {
		t.Errorf("first result = %#v, want %#v", got, want)
	}
	if got, want := results[1].URL, "https://example.org/two?a=1"; got != want {
		t.Errorf("redirect URL = %q, want %q", got, want)
	}
}

func TestWebSearchInvokeUsesFreeProviderAndFormatsResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Query().Get("q"), "current weather"; got != want {
			t.Errorf("query = %q, want %q", got, want)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Error("expected User-Agent header")
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(duckDuckGoResultsFixture))
	}))
	defer server.Close()

	search := &WebSearch{
		client:   &http.Client{Timeout: time.Second},
		endpoint: server.URL + "/html/",
	}
	response, err := search.Invoke(context.Background(), &WebSearchRequest{
		Query: "current weather", MaxResults: 1,
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if !strings.Contains(response.Content, "1. First & Result") ||
		!strings.Contains(response.Content, "URL: https://example.com/one") ||
		strings.Contains(response.Content, "2. Second result") {
		t.Errorf("unexpected formatted response:\n%s", response.Content)
	}
}

func TestWebSearchInvokeRejectsInvalidInput(t *testing.T) {
	search := &WebSearch{client: http.DefaultClient, endpoint: duckDuckGoHTMLSearchURL}
	for _, request := range []*WebSearchRequest{
		{Query: " "},
		{Query: "test", MaxResults: 11},
	} {
		if _, err := search.Invoke(context.Background(), request); err == nil {
			t.Errorf("Invoke(%+v) error = nil, want validation error", request)
		}
	}
}
