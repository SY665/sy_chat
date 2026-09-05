<script setup lang="ts">
import { computed, ref } from 'vue'
import {
    Check,
    LoaderCircle,
    MessageSquare,
    PanelLeftClose,
    PanelLeftOpen,
    Pencil,
    Pin,
    Plus,
    Search,
    Settings,
    Trash2,
    X,
} from '@lucide/vue'

import type { ConversationSummary } from '@/features/conversation/types'
import SettingsDialog from '@/features/settings/components/SettingsDialog.vue'
const props = defineProps<{
    collapsed: boolean
    conversations: ConversationSummary[]
    activeConversationId?: string
    searchQuery: string
    isLoading: boolean
    isCreating: boolean
    isMutating: boolean
    mutatingConversationId: string | null
    error?: string | null
}>()

const emit = defineEmits<{
    toggle: []
    create: []
    select: [id: string]
    rename: [id: string, title: string]
    pin: [id: string, isPinned: boolean]
    search: [query: string]
    remove: [id: string]
}>()

const settingsOpen = ref(false)

const conversationSections = computed(() => [
    {
        label: '置顶',
        conversations: props.conversations.filter(
            (conversation) => conversation.isPinned,
        ),
    },
    {
        label: '最近',
        conversations: props.conversations.filter(
            (conversation) => !conversation.isPinned,
        ),
    },
].filter((section) => section.conversations.length > 0))

const editingID = ref<string | null>(null)
const editingTitle = ref('')

function startRename(conversation: ConversationSummary) {
    editingID.value = conversation.id
    editingTitle.value = conversation.title
}

function cancelRename() {
    editingID.value = null
    editingTitle.value = ''
}

function submitRename() {
    const id = editingID.value
    const title = editingTitle.value.trim()

    if (!id || !title) {
        return
    }

    emit('rename', id, title)
    cancelRename()
}

function handleSearchInput(event: Event) {
    const input = event.target as HTMLInputElement
    emit('search', input.value)
}

function requestRemove(conversation: ConversationSummary) {
    const confirmed = window.confirm(
        `确定删除对话“${conversation.title}”吗？`,
    )

    if (confirmed) {
        emit('remove', conversation.id)
    }
}
</script>

<template>
    <aside
        class="hidden shrink-0 flex-col border-r border-neutral-200 bg-neutral-50 transition-[width] duration-200 dark:border-neutral-800 dark:bg-neutral-900 md:flex"
        :class="collapsed ? 'w-16' : 'w-64'">
        <div class="flex h-14 shrink-0 items-center border-b border-neutral-200 px-3 dark:border-neutral-800"
            :class="collapsed ? 'justify-center' : 'gap-3'">
            <div
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
                <MessageSquare :size="18" />
            </div>

            <span v-if="!collapsed" class="min-w-0 flex-1 truncate text-sm font-semibold">
                SY Chat
            </span>

            <button v-if="!collapsed" type="button"
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-200 hover:text-neutral-950 dark:hover:bg-neutral-800 dark:hover:text-white"
                title="收起侧边栏" @click="emit('toggle')">
                <PanelLeftClose :size="17" />
            </button>
        </div>

        <div class="p-3">
            <button type="button"
                class="flex h-10 w-full items-center rounded-md text-sm transition-colors hover:bg-neutral-200 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-neutral-800"
                :class="collapsed ? 'justify-center' : 'gap-3 px-3'" :disabled="isMutating" title="新对话"
                @click="emit('create')">
                <LoaderCircle v-if="isCreating" class="h-[18px] w-[18px] animate-spin" />
                <Plus v-else :size="18" />

                <span v-if="!collapsed">新对话</span>
            </button>
        </div>

        <div v-if="collapsed" class="px-3">
            <button type="button"
                class="flex h-10 w-full items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-200 hover:text-neutral-950 dark:hover:bg-neutral-800 dark:hover:text-white"
                title="展开侧边栏" @click="emit('toggle')">
                <PanelLeftOpen :size="18" />
            </button>
        </div>

        <div v-if="!collapsed" class="min-h-0 flex-1 overflow-y-auto px-3 py-2">
            <div class="relative mb-2">
                <Search class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-neutral-500"
                    :size="15" />

                <input type="search" :value="searchQuery" placeholder="搜索对话..." aria-label="搜索对话"
                    class="h-9 w-full rounded-md border border-neutral-300 bg-white pl-9 pr-9 text-sm outline-none focus:border-emerald-600 dark:border-neutral-700 dark:bg-neutral-950"
                    @input="handleSearchInput" />

                <button v-if="searchQuery" type="button"
                    class="absolute right-1 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-200 hover:text-neutral-950 dark:hover:bg-neutral-800 dark:hover:text-white"
                    title="清除搜索" aria-label="清除搜索" @click="emit('search', '')">
                    <X :size="14" />
                </button>
            </div>

            <p v-if="searchQuery" class="px-3 pb-2 text-xs font-medium text-neutral-500">
                找到 {{ conversations.length }} 个对话
            </p>

            <div v-if="isLoading" class="flex h-16 items-center justify-center text-neutral-500">
                <LoaderCircle class="h-4 w-4 animate-spin" />
            </div>

            <p v-else-if="error" class="px-3 py-2 text-xs text-red-600 dark:text-red-400">
                {{ error }}
            </p>

            <p v-else-if="conversations.length === 0" class="px-3 py-2 text-xs text-neutral-500">
                {{ searchQuery ? '没有找到匹配的对话' : '暂无对话' }}
            </p>

            <template v-else>
                <section v-for="section in conversationSections" :key="section.label" class="mb-3">
                    <p class="px-3 pb-1 text-xs font-medium text-neutral-500">
                        {{ section.label }}
                    </p>

                    <div v-for="conversation in section.conversations" :key="conversation.id"
                        class="group mb-0.5 flex min-h-9 items-center rounded-md" :class="conversation.id === activeConversationId
                            ? 'bg-neutral-200 dark:bg-neutral-800'
                            : 'hover:bg-neutral-200 dark:hover:bg-neutral-800'
                            ">
                        <form v-if="editingID === conversation.id" class="flex min-w-0 flex-1 items-center gap-1 px-1"
                            @submit.prevent="submitRename">
                            <input v-model="editingTitle" type="text" maxlength="255" autofocus
                                class="h-8 min-w-0 flex-1 rounded-md border border-neutral-400 bg-white px-2 text-sm outline-none focus:border-emerald-600 dark:border-neutral-600 dark:bg-neutral-950"
                                @keydown.esc.prevent="cancelRename" />

                            <button type="submit"
                                class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-emerald-700 hover:bg-neutral-300 dark:text-emerald-400 dark:hover:bg-neutral-700"
                                title="保存标题" aria-label="保存标题">
                                <Check :size="15" />
                            </button>

                            <button type="button"
                                class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-300 dark:hover:bg-neutral-700"
                                title="取消修改" aria-label="取消修改" @click="cancelRename">
                                <X :size="15" />
                            </button>
                        </form>

                        <template v-else>
                            <button type="button" class="min-w-0 flex-1 truncate px-3 py-2 text-left text-sm" :class="{
                                'font-medium': conversation.id === activeConversationId,
                            }" :title="conversation.title" @click="emit('select', conversation.id)">
                                {{ conversation.title }}
                            </button>

                            <LoaderCircle v-if="props.mutatingConversationId === conversation.id"
                                class="mr-2 h-4 w-4 shrink-0 animate-spin text-neutral-500" />

                            <button v-else type="button"
                                class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md hover:bg-neutral-300 hover:text-neutral-950 disabled:opacity-50 dark:hover:bg-neutral-700 dark:hover:text-white"
                                :class="conversation.isPinned
                                    ? 'text-emerald-700 dark:text-emerald-400'
                                    : 'text-neutral-500 opacity-0 group-hover:opacity-100 focus:opacity-100'
                                    " :disabled="props.isMutating" :title="conversation.isPinned ? '取消置顶' : '置顶对话'"
                                :aria-label="conversation.isPinned ? '取消置顶' : '置顶对话'"
                                :aria-pressed="conversation.isPinned"
                                @click="emit('pin', conversation.id, !conversation.isPinned)">
                                <Pin :size="14" :fill="conversation.isPinned ? 'currentColor' : 'none'" />
                            </button>

                            <div v-if="props.mutatingConversationId !== conversation.id"
                                class="flex shrink-0 items-center pr-1 opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
                                <button type="button"
                                    class="flex h-7 w-7 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-300 hover:text-neutral-950 disabled:opacity-50 dark:hover:bg-neutral-700 dark:hover:text-white"
                                    :disabled="props.isMutating" title="修改标题" aria-label="修改标题"
                                    @click="startRename(conversation)">
                                    <Pencil :size="14" />
                                </button>

                                <button type="button"
                                    class="flex h-7 w-7 items-center justify-center rounded-md text-neutral-500 hover:bg-red-100 hover:text-red-700 disabled:opacity-50 dark:hover:bg-red-950 dark:hover:text-red-400"
                                    :disabled="props.isMutating" title="删除对话" aria-label="删除对话"
                                    @click="requestRemove(conversation)">
                                    <Trash2 :size="14" />
                                </button>
                            </div>
                        </template>
                    </div>
                </section>
            </template>
        </div>

        <div class="mt-auto border-t border-neutral-200 p-3 dark:border-neutral-800">
            <button type="button"
                class="flex h-10 w-full items-center rounded-md text-sm transition-colors hover:bg-neutral-200 dark:hover:bg-neutral-800"
                :class="collapsed ? 'justify-center' : 'gap-3 px-3'" title="设置" aria-label="打开设置"
                @click="settingsOpen = true">
                <Settings :size="18" />

                <span v-if="!collapsed">设置</span>
            </button>
        </div>

        <SettingsDialog :open="settingsOpen" @close="settingsOpen = false" />
    </aside>
</template>