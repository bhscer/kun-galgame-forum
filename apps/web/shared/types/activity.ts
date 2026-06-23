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
  user: KunUser
  content: string
  likeCount: number
}

// Rich-card payload for TOPIC_CREATION (BE dto.TopicActivityData). The title
// lives in `content`; this carries the extras the topic feed card shows.
export interface TopicActivityData {
  topicId: number
  excerpt: string
  sections: string[]
  coverImages: string[]
  view: number
  likeCount: number
  favoriteCount: number
  replyCount: number
  commentCount: number
  upvoteTime: Date | string | null
  hasBestAnswer: boolean
  isPoll: boolean
  isNSFW: boolean
  topReply?: ActivityTopReply
  // Reaction counts per key (shared/cacheable). The viewer's own "mine" is
  // hydrated separately via useMyTopicInteractions (the feed is shared-cached).
  reactions: { reaction: string; count: number }[]
}

// Rich-card payload for galgame-scoped activity (BE dto.GalgameActivityData).
// coverHash resolves to a CDN URL via imageTokenUrl. The count fields are only
// populated for the GALGAME_CREATION card (global counts; the viewer's own
// liked/favorited state is NOT carried — it would break the shared feed cache).
export interface GalgameActivityData {
  name: string
  coverHash: string
  language: string
  ageLimit: string
  releaseDate: string | null
  galgameId?: number
  resourceCount?: number
  likeCount?: number
  favoriteCount?: number
  // GALGAME_EDIT only. revisionNumber = the per-galgame revision number the diff
  // endpoint's :rev keys on (used directly); revisionId = the wiki revision ROW
  // id, the legacy fallback resolved id→number for rows synced before the feed
  // carried the number.
  revisionId?: number
  revisionNumber?: number
  // GALGAME_CREATION + GALGAME_EDIT (shared info area), from the wiki detail
  // brief: developer = 制作会社 (officials joined with 、); intro =
  // preferred-language introduction.
  developer?: string
  intro?: string
  // GALGAME_RATING_CREATION only — the rating card's fields.
  rating?: ActivityRatingInfo
  // GALGAME_COMMENT_CREATION — the parent comment being replied to (被评论的评论).
  parentComment?: { content: string }
  // GALGAME_RESOURCE_CREATION — the published resource's spec (download link /
  // 提取码 / 解压码 deliberately omitted).
  resource?: {
    type: string
    language: string
    platform: string
    size: string
    note: string
    likeCount: number
  }
}

// Rating card payload (BE dto.RatingInfo). shortSummary is blank when spoilerLevel
// !== 'none' (the card shows a spoiler notice instead).
export interface ActivityRatingInfo {
  ratingId: number
  overall: number
  playStatus: string
  recommend: string
  shortSummary: string
  spoilerLevel: string
  likeCount: number
  authorId: number
}

// The reply this reply quoted (#floor → its body), shown as a nested block.
export interface ActivityQuotedReply {
  floor: number
  content: string
}

// Rich-card payload for TOPIC_REPLY_CREATION (BE dto.ReplyActivityData). The
// reply body is in `content` (tokens already resolved to @name / #floor).
export interface ReplyActivityData {
  topicTitle: string
  quotedReply?: ActivityQuotedReply
}

// Rich-card payload for TOPIC_COMMENT_CREATION (BE dto.TopicCommentActivityData).
// The comment is on a reply — quotedReply is that reply (被评论的评论), topicTitle
// anchors the bottom; the comment body is in ActivityItem.content.
export interface TopicCommentActivityData {
  topicTitle: string
  quotedReply?: ActivityQuotedReply
}

// Per-type rich-card payload, discriminated by ActivityItem.type. Each card
// casts activity.data to the shape its type carries (the dispatcher routes by
// type, so the cast is safe). Absent for types without a rich card yet.
export type ActivityData =
  | TopicActivityData
  | GalgameActivityData
  | ReplyActivityData
  | TopicCommentActivityData

export interface ActivityItem {
  uniqueId: string
  type: ActivityEventType
  timestamp: Date | string
  actor: KunUser
  link: string
  content: string
  data?: ActivityData
}
