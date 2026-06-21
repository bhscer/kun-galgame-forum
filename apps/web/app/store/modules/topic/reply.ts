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

    // Append a markdown snippet (the @mention + #quote tokens from a 「引用」
    // click) to the draft body, separated by a space. The composer's editor
    // re-syncs from the draft and renders the tokens as chips.
    const appendReply = (markdown: string) => {
      const body = replyDraft.mainContent
      const sep = body && !body.endsWith(' ') ? ' ' : ''
      replyDraft.mainContent = `${body}${sep}${markdown} `
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
