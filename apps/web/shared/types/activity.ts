export type ActivityEventType =
  | 'GALGAME_CREATION'
  | 'GALGAME_COMMENT_CREATION'
  | 'GALGAME_RATING_CREATION'
  | 'GALGAME_RATING_COMMENT_CREATION'
  | 'GALGAME_PR_CREATION'
  | 'GALGAME_EDIT'
  | 'GALGAME_WEBSITE_CREATION'
  | 'GALGAME_WEBSITE_COMMENT_CREATION'
  | 'GALGAME_RESOURCE_CREATION'
  | 'TOOLSET_CREATION'
  | 'TOOLSET_RESOURCE_CREATION'
  | 'TOOLSET_COMMENT_CREATION'
  | 'TOPIC_CREATION'
  | 'TOPIC_REPLY_CREATION'
  | 'TOPIC_COMMENT_CREATION'
  | 'TODO_CREATION'
  | 'UPDATE_LOG_CREATION'
  | 'MESSAGE_UPVOTE'
  | 'MESSAGE_SOLUTION'

// A topic's most-liked reply (excerpt + like count), shown on the topic card.
export interface ActivityTopReply {
  content: string
  likeCount: number
}

// Rich-card payload for TOPIC_CREATION (BE dto.TopicActivityData). The title
// lives in `content`; this carries the extras the topic feed card shows.
export interface TopicActivityData {
  excerpt: string
  sections: string[]
  coverImages: string[]
  view: number
  likeCount: number
  replyCount: number
  commentCount: number
  upvoteTime: Date | string | null
  hasBestAnswer: boolean
  isPoll: boolean
  isNSFW: boolean
  topReply?: ActivityTopReply
}

export interface ActivityItem {
  uniqueId: string
  type: ActivityEventType
  timestamp: Date | string
  actor: KunUser
  link: string
  content: string
  // Per-type rich-card payload (discriminated by `type`); absent for types
  // without a rich card yet. Widen to a union as more types are enriched.
  data?: TopicActivityData
}
