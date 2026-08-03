import { apiRequest } from "./client";

interface ChatRequest{
    message:string
}

export interface ChatResponse{
    reply:string
}

/**
 * 将用户信息发送到go聊天窗口
 */

export function requestChatReply (message:string): Promise<ChatResponse>{
    const body:ChatRequest = {
        message,
    }

    return apiRequest<ChatResponse>('/chat',{
        method:'POST',
        body: JSON.stringify(body)
    })
}