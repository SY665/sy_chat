export type AIProvider = 'local' | 'siliconflow'

export interface AIModel {
    id: string
    name: string
    provider: AIProvider
}

// 对应后端响应中的 data，不包含外层 success 字段。
export interface ModelListData {
    models: AIModel[]
    defaultModelId: string
}