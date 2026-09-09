import { ref } from 'vue'

import { synthesizeSpeech } from '@/lib/api/voice'
import {
    defaultVoiceId,
    isVoiceSupported,
} from '../constants/voices'
import { filterTextForSpeech } from '../utils/filterTextForSpeech'

const selectedVoice = ref(loadSelectedVoice())
const generatingMessageId = ref<string | null>(null)
const playingMessageId = ref<string | null>(null)
const activeMessageId = ref<string | null>(null)
const errorMessageId = ref<string | null>(null)
const playbackError = ref('')

let audio: HTMLAudioElement | null = null
let audioURL: string | null = null
let requestController: AbortController | null = null

function loadSelectedVoice(): string {
    try {
        const savedVoice = localStorage.getItem('sy-chat-selected-voice')
        return savedVoice && isVoiceSupported(savedVoice)
            ? savedVoice
            : defaultVoiceId
    } catch {
        return defaultVoiceId
    }
}

function releaseAudio() {
    if (audio) {
        audio.pause()
        audio.onplay = null
        audio.onpause = null
        audio.onended = null
        audio.onerror = null
        audio = null
    }

    if (audioURL) {
        URL.revokeObjectURL(audioURL)
        audioURL = null
    }

    activeMessageId.value = null
    playingMessageId.value = null
}

function stopSpeech() {
    requestController?.abort()
    requestController = null
    generatingMessageId.value = null
    releaseAudio()
}

function setSelectedVoice(voiceId: string) {
    if (!isVoiceSupported(voiceId)) {
        return
    }

    stopSpeech()
    selectedVoice.value = voiceId

    try {
        localStorage.setItem('sy-chat-selected-voice', voiceId)
    } catch {
        // 浏览器禁用存储时仍保留当前页面内的选择。
    }
}

async function toggleSpeech(messageId: string, content: string) {
    playbackError.value = ''
    errorMessageId.value = null

    if (activeMessageId.value === messageId && audio) {
        if (audio.paused) {
            if (audio.ended) {
                audio.currentTime = 0
            }

            try {
                await audio.play()
            } catch {
                playbackError.value = '无法继续播放音频'
                errorMessageId.value = messageId
            }
        } else {
            audio.pause()
        }
        return
    }

    const filteredText = filterTextForSpeech(content)
    if (!filteredText) {
        playbackError.value = '当前消息没有可朗读的文本'
        errorMessageId.value = messageId
        return
    }

    stopSpeech()

    const controller = new AbortController()
    requestController = controller
    generatingMessageId.value = messageId

    try {
        const blob = await synthesizeSpeech(
            filteredText,
            {
                voice: selectedVoice.value,
                speed: 1,
            },
            controller.signal,
        )

        if (requestController !== controller) {
            return
        }

        audioURL = URL.createObjectURL(blob)
        audio = new Audio(audioURL)
        activeMessageId.value = messageId

        audio.onplay = () => {
            playingMessageId.value = messageId
        }

        audio.onpause = () => {
            if (playingMessageId.value === messageId) {
                playingMessageId.value = null
            }
        }

        audio.onended = () => {
            playingMessageId.value = null
        }

        audio.onerror = () => {
            playingMessageId.value = null
            playbackError.value = '音频播放失败'
            errorMessageId.value = messageId
        }

        await audio.play()
    } catch (error) {
        if (
            !(error instanceof DOMException) ||
            error.name !== 'AbortError'
        ) {
            playbackError.value =
                error instanceof Error
                    ? error.message
                    : '语音生成失败'
            errorMessageId.value = messageId
        }
    } finally {
        if (requestController === controller) {
            requestController = null
            generatingMessageId.value = null
        }
    }
}

export function useSpeechPlayer() {
    return {
        selectedVoice,
        generatingMessageId,
        playingMessageId,
        activeMessageId,
        errorMessageId,
        playbackError,
        setSelectedVoice,
        toggleSpeech,
        stopSpeech,
    }
}