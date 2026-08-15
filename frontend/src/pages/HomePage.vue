<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { MessageSquare } from '@lucide/vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import MainLayout from '@/components/layout/MainLayout.vue'
import ChatInput from '@/features/chat/components/ChatInput.vue'
import MessageList from '@/features/chat/components/MessageList.vue'
import { useChatStore } from '@/features/chat/stores/chat'
import { useConversationStore } from '@/features/conversation/stores/conversation'
import { ApiRequestError } from '@/lib/api/client'

const route = useRoute()
const router = useRouter()
const sidebarCollapsed = ref(false)
const skipNextConversationLoad = ref(false)

const chatStore = useChatStore()
const {
  messages,
  hasMessages,
  isGenerating,
  errorMessage,
} = storeToRefs(chatStore)

const conversationStore = useConversationStore()
const {
  conversations,
  currentConversation,
  isLoading,
  isMutating,
  error: conversationError,
} = storeToRefs(conversationStore)

const conversationId = computed(() => {
  const value = route.params.conversationId
  return typeof value === 'string' ? value : undefined
})

async function loadSelectedConversation(id: string) {
  try {
    const conversation =  await conversationStore.loadConversation(id)

    const persistedMessages = conversation.messages
    .filter(
      (message) =>
        message.role === 'user'||
        message.role === 'assistant',
    )
    .map((message) =>({
      id:message.id,
      role:message.role === 'user' ? 'user' as const : 'assistant' as const,
      content:message.content,
      createdAt:message.createdAt,
    }))

    chatStore.setMessages(persistedMessages)
  } catch (error) {
    if (error instanceof ApiRequestError && error.status === 404) {
      conversationStore.clearCurrentConversation()
      chatStore.clearMessages()
      await router.replace({ name: 'home' })
    }
  }
}

async function handleCreateConversation() {
  try {
    const conversation = await conversationStore.create()
    await router.push({
      name: 'home',
      params: {
        conversationId: conversation.id,
      },
    })
  } catch {
    // 错误信息已经保存在 Conversation Store。
  }
}

async function handleSelectConversation(id: string) {
  if (id === conversationId.value) {
    return
  }

  await router.push({
    name: 'home',
    params: {
      conversationId: id,
    },
  })
}

async function handleRenameConversation(
  id: string,
  title: string,
) {
  try {
    await conversationStore.rename(id, title)
  } catch {
    // Store 已保存错误信息。
  }
}

async function handleRemoveConversation(id: string) {
  try {
    const removingCurrentConversation = id === conversationId.value

    await conversationStore.remove(id)

    if (removingCurrentConversation) {
      chatStore.clearMessages()
      await router.replace('/chat')
    }
  } catch {
    // Store 已保存错误信息。
  }
}

async function handleSend(content: string) {
  let targetConversationID = conversationId.value

  try{
    // 用户在空白 /chat 页面直接发送时，先自动创建对话。
    if(!targetConversationID){
      const conversation = await conversationStore.create()

      skipNextConversationLoad.value = true
      targetConversationID = conversation.id
      
      await router.push({
        name:'home',
        params:{
          conversationId:targetConversationID
        },
      })
    }
    await chatStore.sendMessage(targetConversationID,content)
    await loadSelectedConversation(targetConversationID)
    await conversationStore.loadConversations()
  }catch{
    // Store 已保存错误信息。
  }
}

onMounted(async () => {
  try {
    await conversationStore.loadConversations()

    if (conversationId.value) {
      await loadSelectedConversation(conversationId.value)
    }
  } catch {
    // Store 已保存错误信息。
  }
})

watch(conversationId, async (id, previousID) => {
  if (id === previousID) {
    return
  }

  chatStore.clearMessages()
  conversationStore.clearCurrentConversation()

  if (id) {
    if (skipNextConversationLoad.value) {
    skipNextConversationLoad.value = false
    return
  }

    await loadSelectedConversation(id)
  }
})
</script>

<template>
  <MainLayout>
    <template #sidebar>
      <AppSidebar 
        :collapsed="sidebarCollapsed" 
        :conversations="conversations" 
        :active-conversation-id="conversationId"
        :is-loading="isLoading" 
        :is-creating="isMutating" 
        :error="conversationError"
        @toggle="sidebarCollapsed = !sidebarCollapsed" 
        @create="handleCreateConversation"
        @select="handleSelectConversation" 
        @rename="handleRenameConversation" 
        @remove="handleRemoveConversation" />
    </template>

    <template #header>
      <AppHeader :title="currentConversation?.title ?? '新对话'" />
    </template>

    <section class="flex min-h-0 flex-1 flex-col">
      <div class="min-h-0 flex-1 overflow-y-auto px-4">
        <div v-if="!hasMessages" class="flex min-h-full flex-col items-center justify-center text-center">
          <div
            class="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
            <MessageSquare :size="24" />
          </div>

          <h2 class="text-2xl font-semibold">
            今天有什么可以帮你？
          </h2>
        </div>

        <MessageList v-else :messages="messages" :is-generating="isGenerating" />
      </div>

      <div class="shrink-0 px-4 pb-4 pt-3">
        <div class="mx-auto w-full max-w-3xl">
          <ChatInput :disabled="isGenerating" @send="handleSend" />

          <p v-if="errorMessage" class="mt-2 text-sm text-red-600 dark:text-red-400" role="alert">
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