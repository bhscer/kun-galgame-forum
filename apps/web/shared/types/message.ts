export type MessageType =
  | 'upvoted'
  | 'liked'
  | 'favorite'
  | 'replied'
  | 'solution'
  | 'pin-reply'
  | 'commented'
  | 'expired'
  | 'requested'
  | 'merged'
  | 'declined'
  | 'mentioned'
  | 'admin'

type MessageStatus = 'read' | 'unread'

type MessageSortField = 'time'

export interface MessageRequestData {
  page: string
  limit: string
  type?: MessageType | ''
  sortField?: MessageSortField
  sortOrder: KunOrder
}

export interface Message {
  id: number
  sender: KunUser
  receiverId: number
  link: string
  content: string
  status: MessageStatus
  type: MessageType
  created: Date | string
}

// Wire shape of GET /message (BE dto.MessageListResponse). FE
// pages/message/notice.vue previously referenced an undeclared
// `MessageList` and fell through to TS `any`.
export interface MessageList {
  messages: Message[]
  total: number
}

// System broadcasts use a per-user HWM cursor (system_message_read_state)
// instead of the legacy row-level `status` field. The BE evaluates
// `isRead = id <= cursor` for the caller — see migration 012.
export interface MessageSystemMessage {
  id: number
  isRead: boolean
  content: KunNullable<KunLanguage>
  admin: KunUser
  created: Date | string
}
