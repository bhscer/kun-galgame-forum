export interface ReplyRewriteData {
  id: number
  mainContent: string
}

export interface SuccessfulReplyEvent {
  data: TopicReply
  type: 'created' | 'updated' | 'deleted'
}

export interface ReplyStoreTemp {
  isEdit: boolean
  isScrollToTop: boolean
  scrollToReplyId: number
  isReplyRewriting: boolean
  replyRewrite: ReplyRewriteData | null
  lastSuccessfulReply: SuccessfulReplyEvent | null
}

export interface ReplyStorePersist {
  mode: 'preview' | 'source'
  // A reply is now a single body (inline @mention / #quote tokens replaced the
  // old per-target editors). mainContent is markdown.
  replyDraft: {
    mainContent: string
  }
}
