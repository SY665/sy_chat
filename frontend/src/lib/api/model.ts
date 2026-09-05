import type { ModelListData } from '@/features/model/types'
import { apiRequest } from './client'

export function listModels(): Promise<ModelListData> {
    return apiRequest<ModelListData>('/models')
}