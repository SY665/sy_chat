import { defineStore } from "pinia";

import type { ChatMessage } from "../types";

export const useChatStore = defineStore('chat', {
    state: () => ({
        messages: [] as ChatMessage[],
    }),

    getters: {
        hasMessages: (state) => state.messages.length > 0,
    },

    actions:{
        addUserMessage(content:string){
            const value = content.trim()

            if (!value){
                return
            }

            this.messages.push({
                id:crypto.randomUUID(),
                role:'user',
                content:value,
                createdAt:new Date().toISOString(),
            })
        },

        clearMessages(){
            this.messages = []
        }
    }
})