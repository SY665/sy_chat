package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"sy_chat/internal/models"
	"sy_chat/internal/services/ai"
	toolservice "sy_chat/internal/services/tools"

	"gorm.io/datatypes"
)

const (
	maxToolRounds      = 3
	toolStatusRunning  = "running"
	toolStatusComplete = "complete"
)

var (
	ErrWebSearchUnavailable = errors.New("web search is unavailable")
	ErrToolRoundLimit       = errors.New("AI exceeded the tool round limit")
)

// generationResult 保存最终回复以及本轮产生的完整工具轨迹。
type generationResult struct {
	response    ai.GenerateResponse
	toolCalls   []toolservice.Call
	toolResults []toolservice.Result
}

// encodeToolTrace 将工具调用过程转换为数据库 JSON 字段。
func (result generationResult) encodeToolTrace() (
	datatypes.JSON,
	datatypes.JSON,
	error,
) {
	var callsJSON datatypes.JSON
	var resultsJSON datatypes.JSON

	if len(result.toolCalls) > 0 {
		data, err := json.Marshal(result.toolCalls)
		if err != nil {
			return nil, nil, fmt.Errorf("encode tool calls: %w", err)
		}
		callsJSON = datatypes.JSON(data)
	}

	if len(result.toolResults) > 0 {
		data, err := json.Marshal(result.toolResults)
		if err != nil {
			return nil, nil, fmt.Errorf("encode tool results: %w", err)
		}
		resultsJSON = datatypes.JSON(data)
	}

	return callsJSON, resultsJSON, nil
}

// enabledToolDefinitions 根据服务端能力和用户开关选择本次可用工具。
func (service *Service) enabledToolDefinitions(
	enableWebSearch bool,
) ([]toolservice.Definition, error) {
	definitions := make([]toolservice.Definition, 0, 2)

	if imageTool, exists := service.toolRegistry.Get(
		toolservice.GenerateImageName,
	); exists {
		definitions = append(definitions, imageTool.Definition())
	}

	if !enableWebSearch {
		return definitions, nil
	}

	webSearchTool, exists := service.toolRegistry.Get(
		toolservice.WebSearchName,
	)
	if !exists {
		return nil, ErrWebSearchUnavailable
	}

	definitions = append(definitions, webSearchTool.Definition())
	return definitions, nil
}

func (service *Service) generateWithTools(
	ctx context.Context,
	request ai.GenerateRequest,
	enableWebSearch bool,
) (generationResult, error) {
	definitions, err := service.enabledToolDefinitions(enableWebSearch)
	if err != nil {
		return generationResult{}, err
	}
	if len(definitions) == 0 {
		response, err := service.provider.Generate(ctx, request)
		return generationResult{response: response}, err
	}

	request.Tools = definitions

	messages := append([]ai.Message(nil), request.Messages...)
	toolCalls := make([]toolservice.Call, 0)
	toolResults := make([]toolservice.Result, 0)
	thinkingParts := make([]string, 0)

	for round := 0; round <= maxToolRounds; round++ {
		request.Messages = messages

		response, err := service.provider.Generate(ctx, request)
		if err != nil {
			return generationResult{}, fmt.Errorf(
				"generate AI tool round: %w",
				err,
			)
		}

		if thinking := strings.TrimSpace(response.Thinking); thinking != "" {
			thinkingParts = append(thinkingParts, thinking)
		}

		if len(response.ToolCalls) == 0 {
			response.Thinking = strings.Join(thinkingParts, "\n\n")

			return generationResult{
				response:    response,
				toolCalls:   toolCalls,
				toolResults: toolResults,
			}, nil
		}

		if round == maxToolRounds {
			return generationResult{}, ErrToolRoundLimit
		}

		// 工具调用前的 assistant 消息必须原样回传给模型。
		messages = append(messages, ai.Message{
			Role:             models.MessageRoleAssistant,
			Content:          response.Content,
			ReasoningContent: response.Thinking,
			ToolCalls:        response.ToolCalls,
		})

		for _, call := range response.ToolCalls {
			result, err := service.toolRegistry.Execute(ctx, call)
			if err != nil {
				return generationResult{}, fmt.Errorf(
					"execute AI tool call: %w",
					err,
				)
			}

			toolCalls = append(toolCalls, call)
			toolResults = append(toolResults, result)

			messages = append(messages, ai.Message{
				Role:       models.MessageRoleTool,
				Content:    result.Content,
				ToolCallID: call.ID,
			})
		}
	}

	return generationResult{}, ErrToolRoundLimit
}

// streamWithTools 负责流式生成、执行工具并继续下一轮模型请求。
func (service *Service) streamWithTools(
	ctx context.Context,
	streamingProvider ai.StreamingProvider,
	request ai.GenerateRequest,
	enableWebSearch bool,
	onChunk StreamHandler,
) (generationResult, error) {
	definitions, err := service.enabledToolDefinitions(enableWebSearch)
	if err != nil {
		return generationResult{}, err
	}
	if len(definitions) == 0 {
		err := streamingProvider.Stream(
			ctx,
			request,
			func(chunk ai.StreamChunk) error {
				return onChunk(StreamChunk{
					Content:  chunk.Content,
					Thinking: chunk.Thinking,
				})
			},
		)
		return generationResult{}, err
	}

	request.Tools = definitions

	messages := append([]ai.Message(nil), request.Messages...)
	toolCalls := make([]toolservice.Call, 0)
	toolResults := make([]toolservice.Result, 0)

	for round := 0; round <= maxToolRounds; round++ {
		request.Messages = messages

		var roundContent strings.Builder
		var roundThinking strings.Builder
		roundToolCalls := make([]toolservice.Call, 0)

		err := streamingProvider.Stream(
			ctx,
			request,
			func(chunk ai.StreamChunk) error {
				roundContent.WriteString(chunk.Content)
				roundThinking.WriteString(chunk.Thinking)

				if len(chunk.ToolCalls) > 0 {
					roundToolCalls = append(
						roundToolCalls,
						chunk.ToolCalls...,
					)
				}

				// 工具调用由 Service 内部处理，不直接暴露给文本回调。
				if chunk.Content == "" && chunk.Thinking == "" {
					return nil
				}

				return onChunk(StreamChunk{
					Content:  chunk.Content,
					Thinking: chunk.Thinking,
				})
			},
		)
		if err != nil {
			return generationResult{}, fmt.Errorf(
				"stream AI tool round: %w",
				err,
			)
		}

		if len(roundToolCalls) == 0 {
			return generationResult{
				response: ai.GenerateResponse{
					Content:  roundContent.String(),
					Thinking: roundThinking.String(),
				},
				toolCalls:   toolCalls,
				toolResults: toolResults,
			}, nil
		}

		if round == maxToolRounds {
			return generationResult{}, ErrToolRoundLimit
		}

		// 将模型发起工具调用时的原始消息加入下一轮上下文。
		messages = append(messages, ai.Message{
			Role:             models.MessageRoleAssistant,
			Content:          roundContent.String(),
			ReasoningContent: roundThinking.String(),
			ToolCalls:        roundToolCalls,
		})

		for _, call := range roundToolCalls {
			if err := onChunk(StreamChunk{
				ToolCallID: call.ID,
				ToolName:   call.Function.Name,
				ToolStatus: toolStatusRunning,
			}); err != nil {
				return generationResult{}, fmt.Errorf(
					"report tool start: %w",
					err,
				)
			}

			result, err := service.toolRegistry.Execute(ctx, call)
			if err != nil {
				return generationResult{}, fmt.Errorf(
					"execute streamed AI tool call: %w",
					err,
				)
			}

			toolCalls = append(toolCalls, call)
			toolResults = append(toolResults, result)

			if err := onChunk(StreamChunk{
				ToolCallID: call.ID,
				ToolName:   call.Function.Name,
				ToolStatus: toolStatusComplete,
				Sources:    result.Sources,
				Image:      result.Image,
			}); err != nil {
				return generationResult{}, fmt.Errorf(
					"report tool completion: %w",
					err,
				)
			}

			messages = append(messages, ai.Message{
				Role:       models.MessageRoleTool,
				Content:    result.Content,
				ToolCallID: call.ID,
			})
		}
	}

	return generationResult{}, ErrToolRoundLimit
}
