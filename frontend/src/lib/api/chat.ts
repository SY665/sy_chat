import type { ChatMessage } from '@/features/chat/types'
import type { ApiEnvelope } from './types';

import { apiFetch, apiRequest, ApiRequestError } from "./client";
import { readEventStream } from './sse';

interface ChatRequest {
    conversationId: string
    message: string
    modelId: string
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
 * 将用户信息发送到go聊天窗口
 */
export function requestChatReply(
    conversationId: string,
    message: string,
    modelId = '',
): Promise<ChatResponse> {
    return apiRequest<ChatResponse>('/chat', {
        method: 'POST',
        body: JSON.stringify({
            conversationId,
            message,
            modelId,
        } satisfies ChatRequest),
    })
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
): Promise<ChatResponse> {
    const response = await apiFetch('/chat/stream', {
        method: 'POST',
        signal,
        body: JSON.stringify({
            conversationId,
            message,
            modelId,
        } satisfies ChatRequest),
    })

    if (!response.ok) {
        throw await createResponseError(response)
    }

    const contentType = response.headers.get('Content-Type') ?? ''
    if (!contentType.includes('text/event-stream')) {
        throw new ApiRequestError(
            response.status,
            'INVALID_STREAM_RESPONSE',
            '后端没有返回正确的流式响应',
        )
    }

    const completed: { value?: ChatResponse } = {}

    await readEventStream(response, (event) => {
        if (event.event === 'chunk') {
            const data = parseEventData<ChatChunkData>(event.data)

            if (typeof data.content !== 'string') {
                throw invalidStreamError()
            }

            onChunk(data.content)
            return
        }

        if (event.event === 'done') {
            const data = parseEventData<ChatResponse>(event.data)

            if (!data.userMessage || !data.assistantMessage) {
                throw invalidStreamError()
            }

            completed.value = data
            return
        }

        if (event.event === 'error') {
            const data = parseEventData<ChatStreamError>(event.data)

            throw new ApiRequestError(
                response.status,
                data.code || 'STREAM_FAILED',
                data.message || '生成回复失败',
            )
        }
    })

    if (!completed.value) {
        throw new ApiRequestError(
            response.status,
            'STREAM_INTERRUPTED',
            '流式响应在完成前中断',
        )
    }
    return completed.value
}

function parseEventData<T>(data: string): T {
    try {
        return JSON.parse(data) as T
    } catch {
        throw invalidStreamError()
    }
}

function invalidStreamError(): ApiRequestError {
    return new ApiRequestError(
        200,
        'INVALID_STREAM_EVENT',
        '后端返回了无法解析的流式事件',
    )
}

async function createResponseError(
    response: Response,
): Promise<ApiRequestError> {
    try {
        const payload = (await response.json()) as ApiEnvelope<never>

        return new ApiRequestError(
            response.status,
            payload.error?.code ?? 'REQUEST_FAILED',
            payload.error?.message ?? '请求失败',
        )
    } catch {
        return new ApiRequestError(
            response.status,
            'INVALID_RESPONSE',
            '后端返回了无法解析的数据',
        )
    }
}