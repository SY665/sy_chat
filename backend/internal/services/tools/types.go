package tools

import (
	"context"
	"encoding/json"
)

const (
	ToolTypeFunction  = "function"
	WebSearchName     = "web_search"
	GenerateImageName = "generate_image"
)

// Property 描述工具参数中的一个 JSON Schema 字段。
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// Parameters 描述工具能够接收的参数结构。
type Parameters struct {
	Type                 string              `json:"type"`
	Properties           map[string]Property `json:"properties"`
	Required             []string            `json:"required,omitempty"`
	AdditionalProperties bool                `json:"additionalProperties"`
}

// FunctionDefinition 是发送给 AI Provider 的函数定义。
type FunctionDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

// Definition 使用 OpenAI 兼容的 Function Calling 格式描述工具。
type Definition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionCall 保存模型返回的函数名称和原始 JSON 参数。
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Call 表示 AI 模型发起的一次工具调用。
type Call struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// GeneratedImage 保存生成图片的访问地址和实际尺寸。
type GeneratedImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SearchSource 是返回给前端展示的网页搜索来源。
type SearchSource struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

// Result 同时保存提供给模型的内容和提供给前端的结构化工具结果。
type Result struct {
	ToolCallID string          `json:"toolCallId"`
	Name       string          `json:"name"`
	Content    string          `json:"content"`
	Success    bool            `json:"success"`
	Sources    []SearchSource  `json:"sources,omitempty"`
	Image      *GeneratedImage `json:"image,omitempty"`
}

// Tool 是所有后端工具都需要实现的统一接口。
type Tool interface {
	Definition() Definition
	Execute(ctx context.Context, arguments json.RawMessage) (Result, error)
}
