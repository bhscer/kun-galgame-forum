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
  // GALGAME_EDIT only — wiki revision id for lazily loading the edit diff.
  revisionId?: number
  // GALGAME_CREATION only (from the wiki detail brief): developer = 制作会社
  // (officials joined with 、); intro = preferred-language introduction.
  developer?: string
  intro?: string
  // GALGAME_RATING_CREATION only — the rating card's fields.
  rating?: ActivityRatingInfo
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

// Per-type rich-card payload, discriminated by ActivityItem.type. Each card
// casts activity.data to the shape its type carries (the dispatcher routes by
// type, so the cast is safe). Absent for types without a rich card yet.
export type ActivityData =
  | TopicActivityData
  | GalgameActivityData
  | ReplyActivityData

export interface ActivityItem {
  uniqueId: string
  type: ActivityEventType
  timestamp: Date | string
  actor: KunUser
  link: string
  content: string
  data?: ActivityData
}
