import { defineStore } from "pinia";

import type { ChatMessage,MessageRole } from "../types";
import { ApiRequestError } from "@/lib/api/client";
import { requestChatReply } from "@/lib/api/chat";

function getErrorMessage(error: unknown): string{
    if(error instanceof ApiRequestError){
        return error.message
    }

    return '发送消息失败，请稍后重试'
}

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
        addMessage(role: MessageRole,content:string){
            const value = content.trim()

            if (!value){
                return
            }

            this.messages.push({
                id:crypto.randomUUID(),
                role,
                content:value,
                createdAt:new Date().toISOString(),
            })
        },

        async sendMessage(content: string){
            const value = content.trim()

            if(!value || this.isGenerating){
                return
            }

            this.addMessage('user',value)
            this.errorMessage = null
            this.isGenerating = true

            try{
                //调用后端
                const response = await requestChatReply(value)

                if(!response.reply.trim()){
                    throw new Error('后端返回了空回复')
                }

                this.addMessage('assistant',response.reply)
            }catch(error){
                this.errorMessage = getErrorMessage(error)
            }finally{
                this.isGenerating = false
            }
        },

        clearMessages(){
            this.messages = []
            this.errorMessage = null
        }
    }
})