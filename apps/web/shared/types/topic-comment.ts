export interface TopicComment {
  id: number
  replyId: number
  topicId: number
  user: KunUser
  targetUser: KunUser
  content: string
  isLiked: boolean
  likeCount: number
  created: Date | string
  // set only when the author edits the comment; drives the "(编辑于 …)" hint
  edited?: Date | string | null
}
