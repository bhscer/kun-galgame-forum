import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import type { ReplyStorePersist } from '~/store/types/topic/reply'

export const usePersistKUNGalgameReplyStore = defineStore(
  'KUNGalgameTopicReply',
  () => {
    const mode = ref<ReplyStorePersist['mode']>('preview')
    const replyDraft = reactive<ReplyStorePersist['replyDraft']>({
      mainContent: ''
    })

    // Append the @mention + #quote tokens from a 「引用」 click as a header line,
    // leaving the body to start on the next line (the editor moves the caret
    // there; the user can Backspace to merge). The editor re-syncs from the draft
    // and renders the tokens as chips.
    const appendReply = (markdown: string) => {
      // Dedup: clicking 「引用」 on the same floor repeatedly must not stack the
      // same quote — skip if the draft already references that reply.
      const quoteId = markdown.match(/kungal-reply:(\d+)/)?.[1]
      if (
        quoteId &&
        replyDraft.mainContent.includes(`kungal-reply:${quoteId}`)
      ) {
        return
      }
      const body = replyDraft.mainContent.replace(/\s+$/, '')
      replyDraft.mainContent = body ? `${body}\n\n${markdown}` : markdown
    }

    const resetReplyDraft = () => {
      replyDraft.mainContent = ''
    }

    return {
      mode,
      replyDraft,
      appendReply,
      resetReplyDraft
    }
  },
  {
    persist: true
  }
)
