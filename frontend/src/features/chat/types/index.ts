export type MessageRole = 'user' | 'assistant'
export type MessageStatus =
  | 'sending'
  | 'failed'
  | 'stopped'

export type FileAttachmentType = 'txt' | 'md'

export interface FileAttachment {
  name: string
  type: FileAttachmentType
  size: number
  content: string
}

export interface ChatMessage {
  id: string
  role: MessageRole
  content: string
  attachments?: FileAttachment[]
  thinking?: string
  createdAt: string

  // 只有前端临时消息需要状态，数据库消息不包含该字段。
  status?: MessageStatus
}