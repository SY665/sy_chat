import type {
  ConversationDetail,
  ConversationListData,
  ConversationSummary,
  DeleteConversationData,
  BatchDeleteConversationsData,
  UpdateConversationData,
  UpdateConversationPinnedData,
  PublicShareData,
  ShareConversationData,
  UnshareConversationData,
} from '@/features/conversation/types'

import { apiRequest } from './client'

export function listConversations(): Promise<ConversationListData> {
  return apiRequest<ConversationListData>('/conversations')
}

export function createConversation(
  title = '',
): Promise<ConversationSummary> {
  return apiRequest<ConversationSummary>('/conversations', {
    method: 'POST',
    body: JSON.stringify({ title }),
  })
}

export function getConversation(
  id: string,
): Promise<ConversationDetail> {
  return apiRequest<ConversationDetail>(
    `/conversations/${encodeURIComponent(id)}`,
  )
}

export function updateConversationTitle(
  id: string,
  title: string,
): Promise<UpdateConversationData> {
  return apiRequest<UpdateConversationData>(
    `/conversations/${encodeURIComponent(id)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ title }),
    },
  )
}

export function updateConversationPinned(
  id: string,
  isPinned: boolean,
): Promise<UpdateConversationPinnedData> {
  return apiRequest<UpdateConversationPinnedData>(
    `/conversations/${encodeURIComponent(id)}/pin`,
    {
      method: 'PATCH',
      body: JSON.stringify({ isPinned }),
    },
  )
}

export function deleteConversation(
  id: string,
): Promise<DeleteConversationData> {
  return apiRequest<DeleteConversationData>(
    `/conversations/${encodeURIComponent(id)}`,
    {
      method: 'DELETE',
    },
  )
}

export function deleteConversationsBatch(
  ids: string[],
): Promise<BatchDeleteConversationsData> {
  return apiRequest<BatchDeleteConversationsData>(
    '/conversations/batch',
    {
      method: 'DELETE',
      body: JSON.stringify({ ids }),
    },
  )
}

export function shareConversation(
  id: string,
): Promise<ShareConversationData> {
  return apiRequest<ShareConversationData>(
    `/conversations/${encodeURIComponent(id)}/share`,
    {
      method: 'POST',
    },
  )
}

export function unshareConversation(
  id: string,
): Promise<UnshareConversationData> {
  return apiRequest<UnshareConversationData>(
    `/conversations/${encodeURIComponent(id)}/share`,
    {
      method: 'DELETE',
    },
  )
}

export function getSharedConversation(
  token: string,
): Promise<PublicShareData> {
  return apiRequest<PublicShareData>(
    `/shares/${encodeURIComponent(token)}`,
  )
}

