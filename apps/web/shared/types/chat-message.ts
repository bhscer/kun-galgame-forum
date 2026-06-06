export interface ChatMessageHistoryRequest {
  receiverId: string
  page: string
  limit: string
}

export interface ChatMessageAsideItem {
  chatroomName: string
  content: string
  count: number
  unreadCount: number
  route: string
  title: string
  avatar: string
  // BE returns null/empty when no message has been exchanged yet
  // (NavContactItem.LastMessageTime *string + the nav/system summary
  // emits ""). The Item/SystemItem templates already guard with v-if,
  // so allowing null/"" here removes the formatTimeDifference("") →
  // "NaN years ago" rendering on fresh chat rooms.
  lastMessageTime: Date | string | null
}

export interface ChatMessage {
  id: number
  chatroomName: string
  sender: KunUser
  readBy: KunUser[]
  receiverId: number
  // Raw markdown source (used for editing/quoting). Render contentHtml, not
  // this, in the bubble.
  content: string
  // Server-rendered, sanitized inline+image HTML — safe to v-html.
  contentHtml: string
  isRecall: boolean
  created: Date | string
  recallTime: Date | string | null
  editTime: Date | string | null
}
