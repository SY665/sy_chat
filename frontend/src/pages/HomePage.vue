<script setup lang="ts">
import { ref } from 'vue'
import { MessageSquare } from '@lucide/vue'
import { storeToRefs } from 'pinia'

import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import MainLayout from '@/components/layout/MainLayout.vue'
import ChatInput from '@/features/chat/components/ChatInput.vue'
import { useChatStore } from '@/features/chat/stores/chat'
import MessageList from '@/features/chat/components/MessageList.vue'


const sidebarCollapsed = ref(false)

const chatStore = useChatStore()
const {messages,hasMessages,isGenerating,errorMessage} = storeToRefs(chatStore)

async function handleSend(content: string) {
  await chatStore.sendMessage(content)
}
</script>

<template>
  <MainLayout>
    <template #sidebar>
      <AppSidebar
        :collapsed="sidebarCollapsed"
        @toggle="sidebarCollapsed = !sidebarCollapsed"
      />
    </template>

    <template #header>
      <AppHeader title="新对话" />
    </template>

    <section class="flex min-h-0 flex-1 flex-col">
      <div class="min-h-0 flex-1 overflow-y-auto px-4">
        <div
          v-if="!hasMessages"
          class="flex min-h-full flex-col items-center justify-center text-center"
        >
          <div
            class="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-neutral-950 text-white dark:bg-white dark:text-neutral-950"
          >
            <MessageSquare :size="24" />
          </div>

          <h2 class="text-2xl font-semibold">
            今天有什么可以帮你？
          </h2>
        </div>

        <MessageList 
          v-else
          :messages="messages"
          :is-generating="isGenerating"
        />
      </div>

      <div class="shrink-0 px-4 pb-4 pt-3">
        <div class="mx-auto w-full max-w-3xl">
          <ChatInput 
            :disabled="isGenerating"
            @send="handleSend" 
          />

          <p
           v-if="errorMessage"
           class="mt-2 text-sm text-red-600 dark:text-red-400"
           role="alert">
            {{ errorMessage }}
          </p>

          <p class="mt-3 text-center text-xs text-neutral-500">
            SY Chat 可能会生成不准确的信息，请核查重要内容。
          </p>
        </div>
      </div>
    </section>
  </MainLayout>
</template>