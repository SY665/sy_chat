package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sy_chat/internal/services/imagegen"
	"unicode/utf8"
)

const generateImageMaxPromptLength = 4000

var supportedImageSizes = []string{
	"1024x1024",
	"512x1024",
	"768x512",
	"768x1024",
	"1024x576",
	"576x1024",
}

type generateImageArguments struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt"`
	ImageSize      string `json:"image_size"`
	Seed           *int   `json:"seed"`
}

// GenerateImageTool 将图片生成客户端和本地存储组合成 AI 工具。
type GenerateImageTool struct {
	client  *imagegen.SiliconFlowClient
	storage *imagegen.Storage
}

func NewGenerateImageTool(
	client *imagegen.SiliconFlowClient,
	storage *imagegen.Storage,
) *GenerateImageTool {
	return &GenerateImageTool{
		client:  client,
		storage: storage,
	}
}

func (tool *GenerateImageTool) Definition() Definition {
	return Definition{
		Type: ToolTypeFunction,
		Function: FunctionDefinition{
			Name:        GenerateImageName,
			Description: "生成图片。当用户明确要求创作、绘制或生成图片时调用；每次请求只调用一次。",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"prompt": {
						Type:        "string",
						Description: "详细的英文图片描述，包含主体、风格、光线和构图",
					},
					"negative_prompt": {
						Type:        "string",
						Description: "不希望出现在图片中的内容",
					},
					"image_size": {
						Type:        "string",
						Description: "生成图片的尺寸",
						Enum:        supportedImageSizes,
					},
					"seed": {
						Type:        "integer",
						Description: "可选的随机种子，用于复现相近结果",
					},
				},
				Required:             []string{"prompt"},
				AdditionalProperties: false,
			},
		},
	}
}

func (tool *GenerateImageTool) Execute(
	ctx context.Context,
	arguments json.RawMessage,
) (Result, error) {
	if tool.client == nil || tool.storage == nil {
		return Result{}, fmt.Errorf("image generation tool is not configured")
	}

	var input generateImageArguments
	if err := json.Unmarshal(arguments, &input); err != nil {
		return Result{}, fmt.Errorf(
			"decode image generation arguments: %w",
			err,
		)
	}

	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return Result{}, fmt.Errorf("image prompt is required")
	}
	if utf8.RuneCountInString(prompt) > generateImageMaxPromptLength {
		return Result{}, fmt.Errorf(
			"image prompt exceeds %d characters",
			generateImageMaxPromptLength,
		)
	}

	imageSize := strings.TrimSpace(input.ImageSize)
	if imageSize == "" {
		imageSize = supportedImageSizes[0]
	}

	width, height, err := parseSupportedImageSize(imageSize)
	if err != nil {
		return Result{}, err
	}

	generated, err := tool.client.Generate(ctx, imagegen.GenerateInput{
		Prompt:         prompt,
		NegativePrompt: input.NegativePrompt,
		ImageSize:      imageSize,
		Seed:           input.Seed,
	})
	if err != nil {
		return Result{}, fmt.Errorf("generate image: %w", err)
	}

	stored, err := tool.storage.SaveRemote(ctx, generated.URL)
	if err != nil {
		return Result{}, fmt.Errorf("persist generated image: %w", err)
	}

	return Result{
		Name: GenerateImageName,
		Content: fmt.Sprintf(
			"图片已生成并保存，尺寸为 %s，随机种子为 %d。",
			imageSize,
			generated.Seed,
		),
		Image: &GeneratedImage{
			URL:    stored.URL,
			Width:  width,
			Height: height,
		},
	}, nil
}

func parseSupportedImageSize(value string) (int, int, error) {
	for _, supported := range supportedImageSizes {
		if value != supported {
			continue
		}

		widthText, heightText, _ := strings.Cut(value, "x")
		width, widthErr := strconv.Atoi(widthText)
		height, heightErr := strconv.Atoi(heightText)
		if widthErr != nil || heightErr != nil {
			break
		}

		return width, height, nil
	}

	return 0, 0, fmt.Errorf(
		"unsupported image size: %s",
		value,
	)
}

var _ Tool = (*GenerateImageTool)(nil)
