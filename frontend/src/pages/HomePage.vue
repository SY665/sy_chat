<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { MessageSquare } from '@lucide/vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import MainLayout from '@/components/layout/MainLayout.vue'
import ChatInput from '@/features/chat/components/ChatInput.vue'
import ModelSelect from '@/features/model/components/ModelSelect.vue'
import { useModelStore } from '@/features/model/stores/model'
import MessageList from '@/features/chat/components/MessageList.vue'
import { useChatStore } from '@/features/chat/stores/chat'
import { useConversationStore } from '@/features/conversation/stores/conversation'
import { ApiRequestError } from '@/lib/api/client'
import type { ChatMessage } from '@/features/chat/types'
import { useAppStore } from '@/stores/app'
import { exportConversationAsMarkdown } from '@/features/conversation/utils/exportConversation'

const appStore = useAppStore()
const modelStore = useModelStore()
const route = useRoute()
const router = useRouter()

const { sidebarCollapsed } = storeToRefs(appStore)
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
  filteredConversations,
  searchQuery,
  currentConversation,
  isLoading,
  isCreating,
  isMutating,
  mutatingConversationId,
  error: conversationError,
} = storeToRefs(conversationStore)

const conversationId = computed(() => {
  const value = route.params.conversationId
  return typeof value === 'string' ? value : undefined
})

async function loadSelectedConversation(id: string) {
  try {
    const conversation = await conversationStore.loadConversation(id)
    if (conversationId.value !== id) {
      return
    }

    const persistedMessages = conversation.messages
      .filter(
        (message) =>
          message.role === 'user' ||
          message.role === 'assistant',
      )
      .map((message) => ({
        id: message.id,
        role: message.role === 'user' ? 'user' as const : 'assistant' as const,
        content: message.content,
        createdAt: message.createdAt,
      }))

    chatStore.setMessages(persistedMessages)
  } catch (error) {
    if (conversationId.value !== id) {
      return
    }
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

async function handlePinConversation(
  id: string,
  isPinned: boolean,
) {
  try {
    await conversationStore.setPinned(id, isPinned)
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

function handleExportConversation() {
  if (!hasMessages.value || isGenerating.value) {
    return
  }

  exportConversationAsMarkdown(
    currentConversation.value?.title ?? '新对话',
    messages.value,
  )
}

async function handleSend(content: string) {
  // 在异步创建会话之前记录模型，整个请求使用同一个选择。
  const modelId = modelStore.selectedModelId
  let targetConversationID = conversationId.value

  try {
    // 用户在空白 /chat 页面直接发送时，先自动创建对话。
    if (!targetConversationID) {
      const conversation = await conversationStore.create()

      skipNextConversationLoad.value = true
      targetConversationID = conversation.id

      await router.push({
        name: 'home',
        params: {
          conversationId: targetConversationID
        },
      })
    }
    await chatStore.sendMessage(targetConversationID, content, modelId)

    if (conversationId.value !== targetConversationID) {
      return
    }
    await loadSelectedConversation(targetConversationID)
    await conversationStore.loadConversations()
  } catch {
    // Store 已保存错误信息。
  }
}

async function handleRetry(message: ChatMessage) {
  const canRetry =
    message.status === 'failed' ||
    message.status === 'stopped'

  if (isGenerating.value || !canRetry) {
    return
  }

  // 先移除旧的失败消息，新的发送流程会创建临时消息。
  chatStore.removeMessage(message.id)

  await handleSend(message.content)
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
      <AppSidebar :collapsed="sidebarCollapsed" :conversations="filteredConversations" :search-query="searchQuery"
        :active-conversation-id="conversationId" :mutating-conversation-id="mutatingConversationId"
        :is-loading="isLoading" :is-creating="isCreating" :is-mutating="isMutating" :error="conversationError"
        @toggle="appStore.toggleSidebar" @create="handleCreateConversation" @search="conversationStore.setSearchQuery"
        @select="handleSelectConversation" @rename="handleRenameConversation" @pin="handlePinConversation"
        @remove="handleRemoveConversation" />
    </template>

    <template #header>
      <AppHeader :title="currentConversation?.title ?? '新对话'" :can-export="hasMessages && !isGenerating"
        @export="handleExportConversation" />
    </template>

    <section class="flex min-h-0 flex-1 flex-col">
      <div class="min-h-0 flex-1 overflow-y-auto px-4">
        <div v-if="!hasMessages && !isGenerating"
          class="flex min-h-full flex-col items-center justify-center text-center">
          <div
            class="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
            <MessageSquare :size="24" />
          </div>

          <h2 class="text-2xl font-semibold">
            今天有什么可以帮你？
          </h2>
        </div>

        <MessageList v-else :messages="messages" :is-generating="isGenerating" @retry="handleRetry" />
      </div>

      <div class="shrink-0 px-4 pb-4 pt-3">
        <div class="mx-auto w-full max-w-3xl">
          <ModelSelect class="mb-2" :disabled="isGenerating || isCreating" />
          <ChatInput :is-generating="isGenerating" @send="handleSend" @stop="chatStore.stopGenerating" />

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