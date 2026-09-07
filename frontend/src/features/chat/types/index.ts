export type MessageRole = 'user' | 'assistant'
export type MessageStatus =
  | 'sending'
  | 'failed'
  | 'stopped'

export interface ChatMessage {
  id: string
  role: MessageRole
  content: string
  thinking?: string
  createdAt: string

  // 只有前端临时消息需要状态，数据库消息不包含该字段。
  status?: MessageStatus
}