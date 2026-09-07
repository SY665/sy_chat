package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type siliconFlowStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
	} `json:"choices"`
}

// Stream 请求 SiliconFlow，并逐段转发模型生成的文本。
func (provider *SiliconFlowProvider) Stream(
	ctx context.Context,
	request GenerateRequest,
	onChunk StreamHandler,
) error {
	if provider.apiKey == "" {
		return ErrAPIKeyRequired
	}

	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = provider.model
	}
	if model == "" {
		return ErrModelRequired
	}

	if len(request.Messages) == 0 {
		return ErrNoUserMessage
	}

	if onChunk == nil {
		return ErrStreamHandlerRequired
	}

	requestBody := siliconFlowRequest{
		Model:          model,
		Messages:       request.Messages,
		Stream:         true,
		Temperature:    0.7,
		MaxTokens:      1024,
		EnableThinking: request.EnableThinking,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf(
			"encode SiliconFlow stream request: %w",
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
		return fmt.Errorf(
			"create SiliconFlow stream request: %w",
			err,
		)
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "text/event-stream")
	httpRequest.Header.Set(
		"Authorization",
		"Bearer "+provider.apiKey,
	)

	httpResponse, err := provider.client.Do(httpRequest)
	if err != nil {
		return fmt.Errorf(
			"send SiliconFlow stream request: %w",
			err,
		)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < 200 ||
		httpResponse.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(
			io.LimitReader(httpResponse.Body, 64*1024),
		)

		return fmt.Errorf(
			"SiliconFlow returned status %d: %s",
			httpResponse.StatusCode,
			strings.TrimSpace(string(errorBody)),
		)
	}

	scanner := bufio.NewScanner(httpResponse.Body)

	// 模型事件可能包含较长内容，提高单行扫描上限。
	scanner.Buffer(
		make([]byte, 64*1024),
		1024*1024,
	)

	receivedOutput := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(
			strings.TrimPrefix(line, "data:"),
		)

		if data == "[DONE]" {
			if !receivedOutput {
				return ErrEmptyAIResponse
			}

			return nil
		}

		var chunk siliconFlowStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf(
				"decode SiliconFlow stream chunk: %w",
				err,
			)
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		streamChunk := StreamChunk{
			Content:  chunk.Choices[0].Delta.Content,
			Thinking: chunk.Choices[0].Delta.ReasoningContent,
		}

		if streamChunk.Content == "" && streamChunk.Thinking == "" {
			continue
		}

		receivedOutput = true

		if err := onChunk(streamChunk); err != nil {
			return fmt.Errorf(
				"handle SiliconFlow stream chunk: %w",
				err,
			)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf(
			"read SiliconFlow stream: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	// 没有收到 [DONE] 说明响应可能被截断，不能保存不完整回复。
	return fmt.Errorf(
		"SiliconFlow stream ended before [DONE]: %w",
		io.ErrUnexpectedEOF,
	)
}
