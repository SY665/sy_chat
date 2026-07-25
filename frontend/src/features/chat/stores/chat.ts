import { defineStore } from "pinia";

import type { ChatMessage,MessageRole } from "../types";

export const useChatStore = defineStore('chat', {
    state: () => ({
        messages: [] as ChatMessage[],
        isGenerating:false
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
            this.isGenerating = true

            try{
                //模拟数据
                await new Promise<void>((resolve) =>{
                    window.setTimeout(resolve,800)
                })

                this.addMessage(
                    'assistant',
                    `我收到了你的消息：“${value}”。这是一条临时回复，之后会由 Go 后端生成。`,
                )
            }finally{
                this.isGenerating = false
            }
        },

        clearMessages(){
            this.messages = []
        }
    }
})