import type { FileAttachment } from '@/features/chat/types'

import { apiRequest } from './client'

/**
 * 上传文本文件，并返回后端读取后的附件信息。
 */
export function uploadAttachment(file: File): Promise<FileAttachment> {
  const formData = new FormData()
  formData.append('file', file)

  return apiRequest<FileAttachment>('/attachments', {
    method: 'POST',
    body: formData,
  })
}