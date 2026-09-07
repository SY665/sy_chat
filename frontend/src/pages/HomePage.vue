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
import ThinkingToggle from '@/features/model/components/ThinkingToggle.vue'
import MessageList from '@/features/chat/components/MessageList.vue'
import { useChatStore } from '@/features/chat/stores/chat'
import { useConversationStore } from '@/features/conversation/stores/conversation'
import { ApiRequestError, isAbortError } from '@/lib/api/client'
import type { ChatMessage, FileAttachment, } from '@/features/chat/types'
import { useAppStore } from '@/stores/app'
import { exportConversationAsMarkdown } from '@/features/conversation/utils/exportConversation'
import ShareDialog from '@/features/share/components/ShareDialog.vue'

const appStore = useAppStore()
const modelStore = useModelStore()
const route = useRoute()
const router = useRouter()

const { sidebarCollapsed } = storeToRefs(appStore)
const shareDialogOpen = ref(false)
const skipNextConversationLoad = ref(false)
const preserveMessagesOnNextRouteChange = ref(false)

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
        thinking: message.thinking,
        attachments: message.attachments,
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

async function reloadStoppedConversation(
  conversationID: string,
  previousMessageCount: number,
) {
  const retryDelays = [100, 200, 400, 800]

  // 浏览器会先感知连接取消，后端随后完成持久化，因此进行短暂重试。
  for (const delay of retryDelays) {
    await new Promise<void>((resolve) => {
      window.setTimeout(resolve, delay)
    })

    await loadSelectedConversation(conversationID)

    if (messages.value.length > previousMessageCount) {
      return
    }
  }
}

async function handleSend(content: string, attachments: FileAttachment[] = [],) {
  const previousMessageCount = messages.value.filter(
    (message) => message.status === undefined,
  ).length
  // 在异步创建会话之前记录模型，整个请求使用同一个选择。
  const modelId = modelStore.selectedModelId
  const enableThinking = modelStore.effectiveThinkingEnabled
  let targetConversationID = conversationId.value
  let createdConversationForSend = false

  try {
    // 用户在空白 /chat 页面直接发送时，先自动创建对话。
    if (!targetConversationID) {
      const conversation = await conversationStore.create()

      createdConversationForSend = true
      skipNextConversationLoad.value = true
      targetConversationID = conversation.id

      await router.push({
        name: 'home',
        params: {
          conversationId: targetConversationID
        },
      })
    }
    await chatStore.sendMessage(targetConversationID, content, modelId, enableThinking,attachments,)

    if (conversationId.value !== targetConversationID) {
      return
    }
    await loadSelectedConversation(targetConversationID)
    await conversationStore.loadConversations()
  } catch (error) {
    if (
      isAbortError(error) &&
      targetConversationID &&
      conversationId.value === targetConversationID
    ) {
      // 重新读取后端保存的用户消息和部分 AI 回复，并取得真实 ID。
      await reloadStoppedConversation(
        targetConversationID,
        previousMessageCount,
      )
      await conversationStore.loadConversations()
      return
    }
    if (createdConversationForSend && targetConversationID) {
      try {
        // 首条消息失败时删除刚创建的空会话，但保留失败消息用于重试。
        await conversationStore.remove(targetConversationID)

        if (conversationId.value === targetConversationID) {
          preserveMessagesOnNextRouteChange.value = true
          await router.replace('/chat')
        }
      } catch {
        preserveMessagesOnNextRouteChange.value = false
        // 删除失败时保留当前页面，由 Conversation Store 展示错误。
      }
    }

    // 其他错误信息已经保存在对应 Store。
  }
}

async function regenerateFromUserMessage(
  message: ChatMessage,
  content: string,
) {
  const targetConversationID = conversationId.value

  if (
    isGenerating.value ||
    !targetConversationID ||
    message.role !== 'user' ||
    message.status !== undefined
  ) {
    return
  }

  const modelId = modelStore.selectedModelId
  const enableThinking = modelStore.effectiveThinkingEnabled
  try {
    // 编辑和重新生成共用同一套“截断旧分支并重新发送”流程。
    await chatStore.truncateAndResend(
      targetConversationID,
      message,
      content,
      modelId,
      enableThinking,
    )

    if (conversationId.value !== targetConversationID) {
      return
    }

    await loadSelectedConversation(targetConversationID)
    await conversationStore.loadConversations()
  } catch {
    // Store 已保存截断或重新生成时的错误信息。
  }
}

async function handleRetry(message: ChatMessage) {
  if (isGenerating.value) {
    return
  }

  const isTemporaryMessage =
    message.status === 'failed' ||
    message.status === 'stopped'

  if (isTemporaryMessage) {
    // 临时消息没有数据库 ID，因此直接移除并重新发送。
    chatStore.removeMessage(message.id)
    await handleSend(message.content)
    return
  }

  await regenerateFromUserMessage(message, message.content)
}

async function handleEdit(
  message: ChatMessage,
  content: string,
) {
  await regenerateFromUserMessage(message, content)
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

  if (preserveMessagesOnNextRouteChange.value) {
    preserveMessagesOnNextRouteChange.value = false
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
      <AppHeader :title="currentConversation?.title ?? '新对话'"
        :can-share="!!conversationId && hasMessages && !isGenerating" :can-export="hasMessages && !isGenerating"
        @share="shareDialogOpen = true" @export="handleExportConversation" />
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

        <MessageList v-else :messages="messages" :is-generating="isGenerating" @retry="handleRetry"
          @edit="handleEdit" />
      </div>

      <div class="shrink-0 px-4 pb-4 pt-3">
        <div class="mx-auto w-full max-w-3xl">
          <div class="mb-2 flex min-w-0 items-center justify-between gap-2">
            <ModelSelect full-width class="min-w-0 flex-1" :disabled="isGenerating || isCreating" />

            <ThinkingToggle class="shrink-0" :disabled="isGenerating || isCreating" />
          </div>
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
    <ShareDialog :open="shareDialogOpen" :conversation-id="conversationId ?? ''" @close="shareDialogOpen = false" />
  </MainLayout>
</template>