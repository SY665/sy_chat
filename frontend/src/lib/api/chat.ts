import type { ChatMessage } from '@/features/chat/types'

import { apiRequest } from "./client";

interface ChatRequest{
    conversationId: string
    message:string
}

export interface ChatResponse{
    userMessage: ChatMessage
    assistantMessage: ChatMessage
}

/**
 * 将用户信息发送到go聊天窗口
 */

export function requestChatReply (
    conversationId:string,
    message:string,
): Promise<ChatResponse>{
    const body:ChatRequest = {
        conversationId,
        message,
    }

    return apiRequest<ChatResponse>('/chat',{
        method:'POST',
        body: JSON.stringify(body)
    })
}