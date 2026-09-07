package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sy_chat/internal/models"
	"sy_chat/internal/services/ai"
	"time"
	"unicode/utf8"
)

const (
	MaxMessageLength   = 20000
	maxHistoryMessages = 8
)

var (
	ErrUserIDRequired         = errors.New("user ID is required")
	ErrConversationIDRequired = errors.New("conversation ID is required")
	ErrMessageRequired        = errors.New("message is required")
	ErrMessageTooLong         = errors.New("message is too long")
	ErrStreamingNotSupported  = errors.New("AI provider does not support streaming")
	ErrModelUnavailable       = errors.New("requested model is unavailable")
	ErrMessageIDRequired      = errors.New("message ID is required")
	ErrThinkingNotSupported   = errors.New("thinking mode is not supported by requested model")
)

// Repository 描述聊天业务需要的消息数据库操作。
type Repository interface {
	CreateExchange(
		ctx context.Context,
		userID string,
		title string,
		userMessage *models.Message,
		assistantMessage *models.Message,
	) error

	ListRecentByConversation(
		ctx context.Context,
		conversationID string,
		userID string,
		limit int,
	) ([]models.Message, error)

	DeleteFromUserMessage(
		ctx context.Context,
		conversationID string,
		userID string,
		messageID string,
	) (int64, error)
}

type SendInput struct {
	UserID         string
	ConversationID string
	Content        string
	ModelID        string
	EnableThinking bool
}

// TruncateInput 描述从某条用户消息开始截断对话所需的数据。
type TruncateInput struct {
	UserID         string
	ConversationID string
	MessageID      string
}

type Exchange struct {
	UserMessage      *models.Message
	AssistantMessage *models.Message
}

// StreamChunk 是 Chat Service 向 Handler 暴露的流式内容。
type StreamChunk struct {
	Content  string
	Thinking string
}

type StreamHandler func(chunk StreamChunk) error

type Service struct {
	repository        Repository
	provider          ai.Provider
	defaultModelID    string
	availableModelIDs map[string]struct{}
	thinkingModelIDs  map[string]struct{}
}

func NewService(
	repository Repository,
	provider ai.Provider,
	defaultModelID string,
	availableModelIDs []string,
	thinkingModelIDs []string,
) *Service {
	availableModels := make(map[string]struct{}, len(availableModelIDs)+1)
	for _, modelID := range availableModelIDs {
		availableModels[modelID] = struct{}{}
	}

	// 即使配置列表有误，也保证后端默认模型可以使用。
	availableModels[defaultModelID] = struct{}{}

	thinkingModels := make(map[string]struct{}, len(thinkingModelIDs))
	for _, modelID := range thinkingModelIDs {
		// 只有同时存在于可用模型列表中的模型才能开启思考模式。
		if _, exists := availableModels[modelID]; exists {
			thinkingModels[modelID] = struct{}{}
		}
	}

	return &Service{
		repository:        repository,
		provider:          provider,
		defaultModelID:    defaultModelID,
		availableModelIDs: availableModels,
		thinkingModelIDs:  thinkingModels,
	}
}

// Send 校验消息、生成回复并持久化完整的一轮对话。
func (service *Service) Send(
	ctx context.Context,
	input SendInput,
) (*Exchange, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	conversationID := strings.TrimSpace(input.ConversationID)
	if conversationID == "" {
		return nil, ErrConversationIDRequired
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrMessageRequired
	}

	if utf8.RuneCountInString(content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	// 未指定模型时使用后端默认值，指定值必须在可用范围内。
	modelID, err := service.resolveModelID(input.ModelID)
	if err != nil {
		return nil, err
	}

	if err := service.validateThinkingMode(
		modelID,
		input.EnableThinking,
	); err != nil {
		return nil, err
	}

	history, err := service.repository.ListRecentByConversation(
		ctx,
		conversationID,
		userID,
		maxHistoryMessages,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load conversation context: %w",
			err,
		)
	}

	contextMessages := buildContextMessages(history, content)

	aiResponse, err := service.provider.Generate(
		ctx,
		ai.GenerateRequest{
			Model:          modelID,
			Messages:       contextMessages,
			EnableThinking: input.EnableThinking,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("generate AI reply: %w", err)
	}

	reply := strings.TrimSpace(aiResponse.Content)
	if reply == "" {
		return nil, ai.ErrEmptyAIResponse
	}
	thinking := strings.TrimSpace(aiResponse.Thinking)

	var storedThinking *string

	// 数据库使用 nil 表示该消息没有独立思考内容。
	if thinking != "" {
		storedThinking = &thinking
	}

	now := time.Now()

	userMessage := &models.Message{
		ConversationID: conversationID,
		Role:           models.MessageRoleUser,
		Content:        content,
		CreatedAt:      now,
	}

	assistantMessage := &models.Message{
		ConversationID: conversationID,
		Role:           models.MessageRoleAssistant,
		Content:        reply,
		Thinking:       storedThinking,
		CreatedAt:      now.Add(time.Millisecond),
	}

	title := createConversationTitle(content)

	if err := service.repository.CreateExchange(
		ctx,
		userID,
		title,
		userMessage,
		assistantMessage,
	); err != nil {
		return nil, fmt.Errorf("save chat exchange: %w", err)
	}

	return &Exchange{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
	}, nil
}

// Stream 逐段返回 AI 内容，并在生成完成后保存完整的一轮对话。
func (service *Service) Stream(
	ctx context.Context,
	input SendInput,
	onChunk StreamHandler,
) (*Exchange, error) {
	if onChunk == nil {
		return nil, ai.ErrStreamHandlerRequired
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	conversationID := strings.TrimSpace(input.ConversationID)
	if conversationID == "" {
		return nil, ErrConversationIDRequired
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrMessageRequired
	}

	if utf8.RuneCountInString(content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	// 未指定模型时使用后端默认值，指定值必须在可用范围内。
	modelID, err := service.resolveModelID(input.ModelID)
	if err != nil {
		return nil, err
	}

	if err := service.validateThinkingMode(
		modelID,
		input.EnableThinking,
	); err != nil {
		return nil, err
	}

	history, err := service.repository.ListRecentByConversation(
		ctx,
		conversationID,
		userID,
		maxHistoryMessages,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load conversation context: %w",
			err,
		)
	}

	streamingProvider, ok := service.provider.(ai.StreamingProvider)
	if !ok {
		return nil, ErrStreamingNotSupported
	}

	contextMessages := buildContextMessages(history, content)
	var replyBuilder strings.Builder
	var thinkingBuilder strings.Builder

	err = streamingProvider.Stream(
		ctx,
		ai.GenerateRequest{
			Model:          modelID,
			Messages:       contextMessages,
			EnableThinking: input.EnableThinking,
		},
		func(chunk ai.StreamChunk) error {
			if chunk.Content == "" && chunk.Thinking == "" {
				return nil
			}

			if chunk.Thinking != "" {
				thinkingBuilder.WriteString(chunk.Thinking)
			}

			if chunk.Content != "" {
				replyBuilder.WriteString(chunk.Content)
			}

			return onChunk(StreamChunk{
				Content:  chunk.Content,
				Thinking: chunk.Thinking,
			})
		},
	)
	if err != nil {
		streamCanceled :=
			errors.Is(err, context.Canceled) ||
				errors.Is(ctx.Err(), context.Canceled)

		if streamCanceled {
			partialReply := strings.TrimSpace(replyBuilder.String())
			partialThinking := strings.TrimSpace(thinkingBuilder.String())

			if saveErr := service.saveInterruptedStream(
				userID,
				conversationID,
				content,
				partialReply,
				partialThinking,
			); saveErr != nil {
				return nil, fmt.Errorf(
					"save interrupted chat exchange: %w",
					saveErr,
				)
			}

			return nil, context.Canceled
		}

		return nil, fmt.Errorf("stream AI reply: %w", err)
	}

	reply := strings.TrimSpace(replyBuilder.String())
	if reply == "" {
		return nil, ai.ErrEmptyAIResponse
	}

	thinking := strings.TrimSpace(thinkingBuilder.String())
	var storedThinking *string

	if thinking != "" {
		storedThinking = &thinking
	}

	now := time.Now()

	userMessage := &models.Message{
		ConversationID: conversationID,
		Role:           models.MessageRoleUser,
		Content:        content,
		CreatedAt:      now,
	}

	assistantMessage := &models.Message{
		ConversationID: conversationID,
		Role:           models.MessageRoleAssistant,
		Content:        reply,
		CreatedAt:      now.Add(time.Millisecond),
		Thinking:       storedThinking,
	}

	if err := service.repository.CreateExchange(
		ctx,
		userID,
		createConversationTitle(content),
		userMessage,
		assistantMessage,
	); err != nil {
		return nil, fmt.Errorf(
			"save streamed chat exchange: %w",
			err,
		)
	}

	return &Exchange{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
	}, nil
}

// saveInterruptedStream 使用独立上下文保存被用户停止的流式结果。
// 原 HTTP 请求的 Context 已经取消，不能继续用于数据库操作。
func (service *Service) saveInterruptedStream(
	userID string,
	conversationID string,
	content string,
	partialReply string,
	partialThinking string,
) error {
	saveContext, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	now := time.Now()

	var storedThinking *string

	if partialThinking != "" {
		storedThinking = &partialThinking
	}

	userMessage := &models.Message{
		ConversationID: conversationID,
		Role:           models.MessageRoleUser,
		Content:        content,
		CreatedAt:      now,
	}

	var assistantMessage *models.Message

	// 即使正文尚未开始，只要已有思考内容，也需要保存 AI 消息。
	if partialReply != "" || partialThinking != "" {
		assistantMessage = &models.Message{
			ConversationID: conversationID,
			Role:           models.MessageRoleAssistant,
			Content:        partialReply,
			Thinking:       storedThinking,
			CreatedAt:      now.Add(time.Millisecond),
		}
	}

	return service.repository.CreateExchange(
		saveContext,
		userID,
		createConversationTitle(content),
		userMessage,
		assistantMessage,
	)
}

// TruncateFromUserMessage 删除指定用户消息以及它之后的对话分支。
func (service *Service) TruncateFromUserMessage(
	ctx context.Context,
	input TruncateInput,
) (int64, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return 0, ErrUserIDRequired
	}

	conversationID := strings.TrimSpace(input.ConversationID)
	if conversationID == "" {
		return 0, ErrConversationIDRequired
	}

	messageID := strings.TrimSpace(input.MessageID)
	if messageID == "" {
		return 0, ErrMessageIDRequired
	}

	deletedCount, err := service.repository.DeleteFromUserMessage(
		ctx,
		conversationID,
		userID,
		messageID,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"truncate conversation from user message: %w",
			err,
		)
	}

	return deletedCount, nil
}

// resolveModelID 统一处理普通请求与流式请求的模型选择。
func (service *Service) resolveModelID(requestedModelID string) (string, error) {
	modelID := strings.TrimSpace(requestedModelID)
	if modelID == "" {
		modelID = service.defaultModelID
	}

	if _, exists := service.availableModelIDs[modelID]; !exists {
		return "", ErrModelUnavailable
	}

	return modelID, nil
}

// validateThinkingMode 校验模型是否具有thinking能力
func (service *Service) validateThinkingMode(
	modelID string,
	enableThinking bool,
) error {
	if !enableThinking {
		return nil
	}

	if _, exists := service.thinkingModelIDs[modelID]; !exists {
		return ErrThinkingNotSupported
	}

	return nil
}

// buildContextMessages 将数据库模型转换为 AI Provider 使用的消息格式。
// 工具消息暂未接入，因此这里只传递模型当前能够理解的角色。
func buildContextMessages(
	history []models.Message,
	currentContent string,
) []ai.Message {
	messages := make(
		[]ai.Message,
		0,
		len(history)+1,
	)

	for _, message := range history {
		switch message.Role {
		case models.MessageRoleUser,
			models.MessageRoleAssistant,
			models.MessageRoleSystem:
			messages = append(messages, ai.Message{
				Role:    message.Role,
				Content: message.Content,
			})
		}
	}

	messages = append(messages, ai.Message{
		Role:    models.MessageRoleUser,
		Content: currentContent,
	})

	return messages
}

// createConversationTitle 使用第一条消息生成简短标题。
func createConversationTitle(content string) string {
	const maxTitleLength = 30

	// 去除换行和连续空格，避免标题破坏侧边栏布局。
	normalized := strings.Join(strings.Fields(content), " ")
	characters := []rune(normalized)

	if len(characters) <= maxTitleLength {
		return normalized
	}

	return string(characters[:maxTitleLength]) + "…"
}
