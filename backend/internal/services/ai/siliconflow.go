package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultRequestTimeout = 60 * time.Second

var (
	ErrAPIKeyRequired = errors.New("SiliconFlow API key is required")
	ErrModelRequired  = errors.New("SiliconFlow model is required")
)

type SiliconFlowProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewSiliconFlowProvider(
	apiKey string,
	baseURL string,
	model string,
	timeout time.Duration,
) *SiliconFlowProvider {
	if timeout <= 0 {
		timeout = defaultRequestTimeout
	}

	return &SiliconFlowProvider{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		model:   strings.TrimSpace(model),

		// 外部模型服务可能长时间无响应，因此客户端必须设置超时。
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type siliconFlowRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type siliconFlowResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (provider *SiliconFlowProvider) Generate(
	ctx context.Context,
	request GenerateRequest,
) (GenerateResponse, error) {
	if provider.apiKey == "" {
		return GenerateResponse{}, ErrAPIKeyRequired
	}

	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = provider.model
	}
	if model == "" {
		return GenerateResponse{}, ErrModelRequired
	}

	if len(request.Messages) == 0 {
		return GenerateResponse{}, ErrNoUserMessage
	}

	requestBody := siliconFlowRequest{
		Model:       model,
		Messages:    request.Messages,
		Stream:      false,
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf(
			"encode SiliconFlow request: %w",
			err,
		)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		provider.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf(
			"create SiliconFlow request: %w",
			err,
		)
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set(
		"Authorization",
		"Bearer "+provider.apiKey,
	)

	httpResponse, err := provider.client.Do(httpRequest)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf(
			"send SiliconFlow request: %w",
			err,
		)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < 200 ||
		httpResponse.StatusCode >= 300 {
		// 限制读取长度，防止异常上游返回过大的错误页面。
		errorBody, _ := io.ReadAll(
			io.LimitReader(httpResponse.Body, 64*1024),
		)

		return GenerateResponse{}, fmt.Errorf(
			"SiliconFlow returned status %d: %s",
			httpResponse.StatusCode,
			strings.TrimSpace(string(errorBody)),
		)
	}

	var responseBody siliconFlowResponse
	if err := json.NewDecoder(httpResponse.Body).
		Decode(&responseBody); err != nil {
		return GenerateResponse{}, fmt.Errorf(
			"decode SiliconFlow response: %w",
			err,
		)
	}

	if len(responseBody.Choices) == 0 {
		return GenerateResponse{}, ErrEmptyAIResponse
	}

	content := strings.TrimSpace(
		responseBody.Choices[0].Message.Content,
	)
	if content == "" {
		return GenerateResponse{}, ErrEmptyAIResponse
	}

	return GenerateResponse{
		Content: content,
	}, nil
}
