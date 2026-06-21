import type { TopicComment } from './topic-comment'

export interface TopicReply {
  id: number
  topicId: number
  floor: number
  user: KunUser & { moemoepoint: number }
  contentMarkdown: string
  contentHtml: string

  likeCount: number
  isLiked: boolean
  dislikeCount: number
  isDisliked: boolean

  comment: TopicComment[]
  created: Date | string
  edited: Date | string | null

  isPinned: boolean
  isBestAnswer: boolean
}

