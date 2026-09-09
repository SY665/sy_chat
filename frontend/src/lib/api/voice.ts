import { ApiRequestError, apiFetch, apiRequest } from './client'
import type { ApiEnvelope } from './types'

export interface SpeechResult {
    text: string
}

function getAudioExtension(mimeType: string): string {
    if (mimeType.includes('mp4')) {
        return 'm4a'
    }

    if (mimeType.includes('ogg')) {
        return 'ogg'
    }

    return 'webm'
}

/**
 * 将浏览器录音上传到后端进行语音识别。
 */
export function transcribeSpeech(
    audio: Blob,
    signal?: AbortSignal,
): Promise<SpeechResult> {
    const formData = new FormData()
    const extension = getAudioExtension(audio.type)

    formData.append(
        'audio',
        new File([audio], `recording.${extension}`, {
            type: audio.type,
        }),
    )

    return apiRequest<SpeechResult>('/speech', {
        method: 'POST',
        body: formData,
        signal,
    })
}

export interface TextToSpeechOptions {
    voice: string
    speed?: number
}

/**
 * 请求后端生成 MP3 音频。
 * TTS 成功响应是二进制内容，因此不能使用只处理 JSON 的 apiRequest。
 */
export async function synthesizeSpeech(
    text: string,
    options: TextToSpeechOptions,
    signal?: AbortSignal,
): Promise<Blob> {
    const response = await apiFetch('/tts', {
        method: 'POST',
        body: JSON.stringify({
            text,
            voice: options.voice,
            speed: options.speed ?? 1,
        }),
        signal,
    })

    if (!response.ok) {
        const payload = await response
            .json()
            .catch(() => null) as ApiEnvelope<never> | null

        throw new ApiRequestError(
            response.status,
            payload?.error?.code ?? 'SPEECH_SYNTHESIS_FAILED',
            payload?.error?.message ?? '语音生成失败',
            response.headers.get('X-Request-ID') ?? '',
        )
    }

    const audio = await response.blob()
    if (audio.size === 0) {
        throw new ApiRequestError(
            response.status,
            'EMPTY_AUDIO_RESPONSE',
            '后端返回了空音频',
        )
    }

    return audio
}