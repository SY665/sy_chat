import type { ChatMessage } from '@/features/chat/types'
import type { ApiEnvelope } from './types';

import { apiFetch, apiRequest, ApiRequestError } from "./client";
import { readEventStream } from './sse';

interface ChatRequest {
    conversationId: string
    message: string
    modelId: string
    enableThinking: boolean
}

interface ChatChunkData {
    content: string
}

interface ChatStreamError {
    code: string
    message: string
}

export interface ChatResponse {
    userMessage: ChatMessage
    assistantMessage: ChatMessage
}

/**
 * 对应后端 TruncateMessagesData。
 */
export interface TruncateMessagesResponse {
    conversationId: string
    messageId: string
    deletedCount: number
}

/**
 * 将用户信息发送到go聊天窗口
 */
export function requestChatReply(
    conversationId: string,
    message: string,
    modelId = '',
    enableThinking = false,
): Promise<ChatResponse> {
    return apiRequest<ChatResponse>('/chat', {
        method: 'POST',
        body: JSON.stringify({
            conversationId,
            message,
            modelId,
            enableThinking,
        } satisfies ChatRequest),
    })
}

/**
 * 删除目标用户消息以及它之后的全部消息。
 */
export function truncateMessagesFromUserMessage(
    conversationId: string,
    messageId: string,
): Promise<TruncateMessagesResponse> {
    const encodedConversationId = encodeURIComponent(conversationId)
    const encodedMessageId = encodeURIComponent(messageId)

    return apiRequest<TruncateMessagesResponse>(
        `/conversations/${encodedConversationId}/messages/${encodedMessageId}/tail`,
        {
            method: 'DELETE',
        },
    )
}

/**
 * 发送聊天消息并持续接收 AI 增量文本。
 */
export async function requestChatStream(
    conversationId: string,
    message: string,
    onChunk: (content: string) => void,
    signal?: AbortSignal,
    modelId = '',
    onThinking?: (content: string) => void,
    enableThinking = false,
): Promise<ChatResponse> {
    const response = await apiFetch('/chat/stream', {
        method: 'POST',
        signal,
        body: JSON.stringify({
            conversationId,
            message,
            modelId,
            enableThinking,
        } satisfies ChatRequest),
    })

    const requestId = response.headers.get('X-Request-ID') ?? ''

    if (!response.ok) {
        throw await createResponseError(response)
    }

    const contentType = response.headers.get('Content-Type') ?? ''
    if (!contentType.includes('text/event-stream')) {
        throw new ApiRequestError(
            response.status,
            'INVALID_STREAM_RESPONSE',
            '后端没有返回正确的流式响应',
            requestId,
        )
    }

    const completed: { value?: ChatResponse } = {}

    await readEventStream(response, (event) => {
        if (event.event === 'thinking') {
            const data = parseEventData<ChatChunkData>(
                event.data,
                requestId,
            )

            if (typeof data.content !== 'string') {
                throw invalidStreamError(requestId)
            }

            onThinking?.(data.content)
            return
        }

        if (event.event === 'chunk') {
            const data = parseEventData<ChatChunkData>(event.data, requestId)

            if (typeof data.content !== 'string') {
                throw invalidStreamError(requestId)
            }

            onChunk(data.content)
            return
        }

        if (event.event === 'done') {
            const data = parseEventData<ChatResponse>(event.data, requestId)

            if (!data.userMessage || !data.assistantMessage) {
                throw invalidStreamError(requestId)
            }

            completed.value = data
            return
        }

        if (event.event === 'error') {
            const data = parseEventData<ChatStreamError>(event.data, requestId)

            throw new ApiRequestError(
                response.status,
                data.code || 'STREAM_FAILED',
                data.message || '生成回复失败',
                requestId,
            )
        }
    })

    if (!completed.value) {
        throw new ApiRequestError(
            response.status,
            'STREAM_INTERRUPTED',
            '流式响应在完成前中断',
            requestId,
        )
    }
    return completed.value
}

function parseEventData<T>(data: string, requestId: string): T {
    try {
        return JSON.parse(data) as T
    } catch {
        throw invalidStreamError(requestId)
    }
}

function invalidStreamError(requestId = ''): ApiRequestError {
    return new ApiRequestError(
        200,
        'INVALID_STREAM_EVENT',
        '后端返回了无法解析的流式事件',
        requestId,
    )
}

async function createResponseError(
    response: Response,
): Promise<ApiRequestError> {
    const requestId = response.headers.get('X-Request-ID') ?? ''
    try {
        const payload = (await response.json()) as ApiEnvelope<never>

        return new ApiRequestError(
            response.status,
            payload.error?.code ?? 'REQUEST_FAILED',
            payload.error?.message ?? '请求失败',
            requestId,
        )
    } catch {
        return new ApiRequestError(
            response.status,
            'INVALID_RESPONSE',
            '后端返回了无法解析的数据',
            requestId,
        )
    }
}