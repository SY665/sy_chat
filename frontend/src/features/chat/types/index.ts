export type MessageRole = 'user' | 'assistant'
export type MessageStatus =
  | 'sending'
  | 'failed'
  | 'stopped'

export type ToolStatus = 'running' | 'complete'

export interface SearchSource {
  title: string
  url: string
  snippet?: string
}

export interface GeneratedImage {
  url: string
  width: number
  height: number
}

export interface ToolStreamEvent {
  toolCallId: string
  name: string
  status: ToolStatus
  sources: SearchSource[]
  image?: GeneratedImage
}

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
  toolEvents?: ToolStreamEvent[]
  createdAt: string

  // 只有前端临时消息需要状态，数据库消息不包含该字段。
  status?: MessageStatus
}