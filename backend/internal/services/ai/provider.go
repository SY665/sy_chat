package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoUserMessage         = errors.New("AI request contains no user message")
	ErrEmptyAIResponse       = errors.New("AI returned an empty response")
	ErrStreamHandlerRequired = errors.New("AI stream handler is required")
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateRequest struct {
	Model          string
	Messages       []Message
	EnableThinking bool
}

type GenerateResponse struct {
	Content  string
	Thinking string
}

type Provider interface {
	Generate(
		ctx context.Context,
		request GenerateRequest,
	) (GenerateResponse, error)
}

// StreamChunk 表示模型一次返回的增量内容。
type StreamChunk struct {
	Content  string
	Thinking string
}

// StreamHandler 接收模型每次生成的增量内容。
type StreamHandler func(chunk StreamChunk) error

// StreamingProvider 描述支持流式生成的 AI Provider。
// 它保留普通 Provider 能力，便于流式接口失败时复用非流式逻辑。
type StreamingProvider interface {
	Provider

	Stream(
		ctx context.Context,
		request GenerateRequest,
		onChunk StreamHandler,
	) error
}

// LocalProvider 用于开发阶段，不发送任何外部网络请求。
type LocalProvider struct{}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}

func (*LocalProvider) Generate(
	ctx context.Context,
	request GenerateRequest,
) (GenerateResponse, error) {
	// 在生成回复前响应请求取消，避免已经断开的请求继续工作。
	if err := ctx.Err(); err != nil {
		return GenerateResponse{}, err
	}

	for index := len(request.Messages) - 1; index >= 0; index-- {
		message := request.Messages[index]

		if message.Role != "user" {
			continue
		}

		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}

		return GenerateResponse{
			Content: fmt.Sprintf(
				"我收到了你的消息：“%s”。这条回复来自 Go 后端。",
				content,
			),
		}, nil
	}

	return GenerateResponse{}, ErrNoUserMessage
}

// Stream 将本地完整回复拆分成较小片段，模拟模型流式输出。
func (provider *LocalProvider) Stream(
	ctx context.Context,
	request GenerateRequest,
	onChunk StreamHandler,
) error {
	if onChunk == nil {
		return ErrStreamHandlerRequired
	}

	response, err := provider.Generate(ctx, request)
	if err != nil {
		return err
	}

	// 使用 rune 切分，避免从 UTF-8 字节中间截断中文字符。
	characters := []rune(response.Content)
	const chunkSize = 6

	for start := 0; start < len(characters); start += chunkSize {
		if err := ctx.Err(); err != nil {
			return err
		}

		end := start + chunkSize
		if end > len(characters) {
			end = len(characters)
		}

		if err := onChunk(StreamChunk{
			Content: string(characters[start:end]),
		}); err != nil {
			return fmt.Errorf(
				"handle local stream chunk: %w",
				err,
			)
		}
	}

	return nil
}
