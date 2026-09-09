package imagegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultImageSize               = "1024x1024"
	maxImageGenerationResponseSize = 1 << 20
)

// GenerateInput 描述一次图片生成请求。
type GenerateInput struct {
	Prompt         string
	NegativePrompt string
	ImageSize      string
	Seed           *int
}

// GenerateResult 保存 SiliconFlow 返回的临时图片信息。
type GenerateResult struct {
	URL  string
	Seed int
}

type siliconFlowGenerateRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	ImageSize      string `json:"image_size"`
	Seed           *int   `json:"seed,omitempty"`
}

type siliconFlowGeneratedImage struct {
	URL string `json:"url"`
}

type siliconFlowGenerateResponse struct {
	Images []siliconFlowGeneratedImage `json:"images"`
	Seed   int                         `json:"seed"`
}

// SiliconFlowClient 负责调用 SiliconFlow 图片生成接口。
type SiliconFlowClient struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
}

func NewSiliconFlowClient(
	apiKey string,
	baseURL string,
	model string,
	timeout time.Duration,
) *SiliconFlowClient {
	return &SiliconFlowClient{
		apiKey:   strings.TrimSpace(apiKey),
		endpoint: strings.TrimRight(baseURL, "/") + "/images/generations",
		model:    strings.TrimSpace(model),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// newGenerateRequest 校验输入并构造 SiliconFlow 图片生成请求。
func (client *SiliconFlowClient) newGenerateRequest(
	ctx context.Context,
	input GenerateInput,
) (*http.Request, error) {
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("image prompt is required")
	}
	if client.apiKey == "" {
		return nil, fmt.Errorf("SiliconFlow API key is not configured")
	}
	if client.model == "" {
		return nil, fmt.Errorf("image generation model is not configured")
	}

	imageSize := strings.TrimSpace(input.ImageSize)
	if imageSize == "" {
		imageSize = defaultImageSize
	}

	body, err := json.Marshal(siliconFlowGenerateRequest{
		Model:          client.model,
		Prompt:         prompt,
		NegativePrompt: strings.TrimSpace(input.NegativePrompt),
		ImageSize:      imageSize,
		Seed:           input.Seed,
	})
	if err != nil {
		return nil, fmt.Errorf("encode image generation request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		client.endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create image generation request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+client.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	return request, nil
}

// Generate 请求 SiliconFlow 生成图片并返回临时图片地址。
func (client *SiliconFlowClient) Generate(
	ctx context.Context,
	input GenerateInput,
) (GenerateResult, error) {
	request, err := client.newGenerateRequest(ctx, input)
	if err != nil {
		return GenerateResult{}, err
	}

	response, err := client.client.Do(request)
	if err != nil {
		return GenerateResult{}, fmt.Errorf(
			"send image generation request: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return GenerateResult{}, fmt.Errorf(
			"SiliconFlow image API returned status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(errorBody)),
		)
	}

	responseBody, err := io.ReadAll(io.LimitReader(
		response.Body,
		maxImageGenerationResponseSize+1,
	))
	if err != nil {
		return GenerateResult{}, fmt.Errorf(
			"read image generation response: %w",
			err,
		)
	}
	if len(responseBody) > maxImageGenerationResponseSize {
		return GenerateResult{}, fmt.Errorf(
			"image generation response exceeds size limit",
		)
	}

	var payload siliconFlowGenerateResponse
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return GenerateResult{}, fmt.Errorf(
			"decode image generation response: %w",
			err,
		)
	}
	if len(payload.Images) == 0 {
		return GenerateResult{}, fmt.Errorf(
			"SiliconFlow image API returned no images",
		)
	}

	imageURL := strings.TrimSpace(payload.Images[0].URL)
	if imageURL == "" {
		return GenerateResult{}, fmt.Errorf(
			"SiliconFlow image API returned an empty image URL",
		)
	}

	return GenerateResult{
		URL:  imageURL,
		Seed: payload.Seed,
	}, nil
}
