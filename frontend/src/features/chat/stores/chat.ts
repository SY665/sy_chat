import { defineStore } from "pinia";

import type { ChatMessage} from "../types";
import { ApiRequestError } from "@/lib/api/client";
import { requestChatReply } from "@/lib/api/chat";

export const useChatStore = defineStore('chat', {
    state: () => ({
        messages: [] as ChatMessage[],
        isGenerating:false,
        errorMessage:null as string | null
    }),

    getters: {
        hasMessages: (state) => state.messages.length > 0,
    },

    actions:{
        async sendMessage(
            conversationID:string,
            content: string,
        ){
            const value = content.trim()

            if(!value || this.isGenerating || !conversationID){
                return
            }

            this.errorMessage = null
            this.isGenerating = true

            try{
                //调用后端
                const response = await requestChatReply(
                    conversationID,
                    value,
                )

                // 使用数据库生成的真实 ID 和时间，不再使用前端临时 ID。
                this.messages.push(
                    response.userMessage,
                    response.assistantMessage,
                )
            }catch(error){
                this.errorMessage = getErrorMessage(error)
                throw error
            }finally{
                this.isGenerating = false
            }
        },

        setMessages(messages: ChatMessage[]){
            this.messages = messages
            this.errorMessage = null
        },

        clearMessages(){
            this.messages = []
            this.errorMessage = null
        },
    },
})

function getErrorMessage(error: unknown): string{
    if(error instanceof ApiRequestError){
        return error.message
    }

    return '发送消息失败，请稍后重试'
}