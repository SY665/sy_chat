package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	webSearchDepth            = "basic"
	webSearchMaxResults       = 6
	webSearchMaxSources       = 5
	webSearchMaxQueryLength   = 500
	webSearchMaxResponseBytes = 2 << 20
	webSearchSnippetLength    = 300
)

type webSearchArguments struct {
	Query string `json:"query"`
}

type tavilySearchRequest struct {
	Query         string `json:"query"`
	SearchDepth   string `json:"search_depth"`
	IncludeAnswer bool   `json:"include_answer"`
	MaxResults    int    `json:"max_results"`
}

type tavilySearchResponse struct {
	Answer  string               `json:"answer"`
	Results []tavilySearchResult `json:"results"`
}

type tavilySearchResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// WebSearchTool 使用 Tavily Search API 查询实时网页内容。
type WebSearchTool struct {
	apiKey    string
	searchURL string
	client    *http.Client
}

func NewWebSearchTool(
	apiKey string,
	searchURL string,
	timeout time.Duration,
) *WebSearchTool {
	return &WebSearchTool{
		apiKey:    apiKey,
		searchURL: searchURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (tool *WebSearchTool) Definition() Definition {
	return Definition{
		Type: ToolTypeFunction,
		Function: FunctionDefinition{
			Name:        WebSearchName,
			Description: "搜索互联网以获取实时信息、新闻、最新数据和当前事件。",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"query": {
						Type:        "string",
						Description: "需要搜索的准确关键词或问题",
					},
				},
				Required:             []string{"query"},
				AdditionalProperties: false,
			},
		},
	}
}

var _ Tool = (*WebSearchTool)(nil)

func (tool *WebSearchTool) Execute(
	ctx context.Context,
	arguments json.RawMessage,
) (Result, error) {
	var input webSearchArguments
	if err := json.Unmarshal(arguments, &input); err != nil {
		return Result{}, fmt.Errorf("decode web search arguments: %w", err)
	}

	query := strings.TrimSpace(input.Query)
	if query == "" {
		return Result{}, fmt.Errorf("web search query is required")
	}
	if utf8.RuneCountInString(query) > webSearchMaxQueryLength {
		return Result{}, fmt.Errorf(
			"web search query exceeds %d characters",
			webSearchMaxQueryLength,
		)
	}
	if strings.TrimSpace(tool.apiKey) == "" {
		return Result{}, fmt.Errorf("Tavily API key is not configured")
	}

	body, err := json.Marshal(tavilySearchRequest{
		Query:         query,
		SearchDepth:   webSearchDepth,
		IncludeAnswer: true,
		MaxResults:    webSearchMaxResults,
	})
	if err != nil {
		return Result{}, fmt.Errorf("encode Tavily request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tool.searchURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return Result{}, fmt.Errorf("create Tavily request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+tool.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := tool.client.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf("send Tavily request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return Result{}, fmt.Errorf(
			"Tavily returned status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(errorBody)),
		)
	}

	var payload tavilySearchResponse
	reader := io.LimitReader(response.Body, webSearchMaxResponseBytes)
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		return Result{}, fmt.Errorf("decode Tavily response: %w", err)
	}

	return formatWebSearchResult(payload), nil
}

func formatWebSearchResult(payload tavilySearchResponse) Result {
	parts := make([]string, 0, len(payload.Results)+1)
	sources := make([]SearchSource, 0, webSearchMaxSources)

	if answer := strings.TrimSpace(payload.Answer); answer != "" {
		parts = append(parts, "搜索摘要：\n"+answer)
	}

	for _, item := range payload.Results {
		title := strings.TrimSpace(item.Title)
		url := strings.TrimSpace(item.URL)
		if title == "" || url == "" {
			continue
		}

		snippet := truncateSearchText(item.Content, webSearchSnippetLength)
		parts = append(parts, fmt.Sprintf(
			"%d. %s\n来源：%s\n%s",
			len(sources)+1,
			title,
			url,
			snippet,
		))

		sources = append(sources, SearchSource{
			Title:   title,
			URL:     url,
			Snippet: snippet,
		})

		if len(sources) >= webSearchMaxSources {
			break
		}
	}

	if len(parts) == 0 {
		parts = append(parts, "未找到相关搜索结果")
	}

	return Result{
		Name:    WebSearchName,
		Content: strings.Join(parts, "\n\n"),
		Success: true,
		Sources: sources,
	}
}

func truncateSearchText(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}

	return string(runes[:limit]) + "..."
}
