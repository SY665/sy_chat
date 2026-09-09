export interface VoiceOption {
    id: string
    label: string
    gender: 'female' | 'male'
    description: string
}

export const voiceOptions: VoiceOption[] = [
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:diana',
        label: 'Diana',
        gender: 'female',
        description: '欢快女声',
    },
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:claire',
        label: 'Claire',
        gender: 'female',
        description: '温柔女声',
    },
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:anna',
        label: 'Anna',
        gender: 'female',
        description: '沉稳女声',
    },
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:bella',
        label: 'Bella',
        gender: 'female',
        description: '激情女声',
    },
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:alex',
        label: 'Alex',
        gender: 'male',
        description: '沉稳男声',
    },
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:david',
        label: 'David',
        gender: 'male',
        description: '欢快男声',
    },
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:charles',
        label: 'Charles',
        gender: 'male',
        description: '磁性男声',
    },
    {
        id: 'FunAudioLLM/CosyVoice2-0.5B:benjamin',
        label: 'Benjamin',
        gender: 'male',
        description: '低沉男声',
    },
]

export const defaultVoiceId = voiceOptions[0]!.id

export function isVoiceSupported(voiceId: string): boolean {
    return voiceOptions.some((voice) => voice.id === voiceId)
}