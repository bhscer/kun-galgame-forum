import type { HomeTopic, HomeGalgame } from './home'

export type SearchResultTopic = HomeTopic
export type SearchResultGalgame = HomeGalgame

// BE `UserItem` (apps/api/internal/search/dto) currently leaves
// `moemoepoint` and `created` zero — they're not joined from
// kungal_user_state at search time. Type them as optional so the FE
// template can guard with v-if rather than render `0` / Date(0).
export interface SearchResultUser extends KunUser {
  bio: string
  moemoepoint?: number
  created?: Date | string
}

export interface SearchResultReplyTarget {
  id: number
  user: KunUser
  content: string
  contentPreview: string
}

// `targets` is reserved for future "show what this reply is quoting".
// BE `ReplyItem` doesn't populate it today, so FE marks it optional;
// the consuming card guards with v-if to avoid empty-block layout gaps.
export interface SearchResultReply {
  topicId: number
  topicTitle: string
  floor: number
  content: string
  user: KunUser
  targets?: SearchResultReplyTarget[]
  created: Date | string
}

// `targetUser` is the "comment-chain parent" — BE `CommentItem`
// doesn't carry it today; FE guards rendering with v-if.
export type SearchResultComment = {
  topicId: number
  topicTitle: string
  content: string
  user: KunUser
  targetUser?: KunUser
  created: Date | string
}

export type SearchType = 'topic' | 'galgame' | 'user' | 'reply' | 'comment'
export type SearchResult =
  | SearchResultTopic
  | SearchResultGalgame
  | SearchResultUser
  | SearchResultReply
  | SearchResultComment
