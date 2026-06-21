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

// A floor the draft is replying to. Shown as a removable chip above the editor;
// its @mention (notify) + #quote (link) tokens are prepended to the body on
// publish. Kept OUT of the editor content so the composer is a plain empty
// editor with native caret behavior (no fragile inline-header cursor placement).
export interface ReplyReference {
  userId: number
  userName: string
  replyId: number
  floor: number
}

export interface ReplyStorePersist {
  mode: 'preview' | 'source'
  // A reply is now a single body (inline @mention / #quote tokens replaced the
  // old per-target editors). mainContent is markdown.
  replyDraft: {
    mainContent: string
  }
  replyReferences: ReplyReference[]
}
