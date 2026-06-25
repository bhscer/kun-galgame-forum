export interface TopicComment {
  id: number
  replyId: number
  topicId: number
  // The comment this one replies to (nested comments); null/undefined = a
  // top-level comment attached to the reply directly.
  parentCommentId?: number | null
  user: KunUser
  targetUser: KunUser
  content: string
  isLiked: boolean
  likeCount: number
  created: Date | string
  // set only when the author edits the comment; drives the "(编辑于 …)" hint
  edited?: Date | string | null
}
