import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import type {
  ReplyReference,
  ReplyStorePersist
} from '~/store/types/topic/reply'

export const usePersistKUNGalgameReplyStore = defineStore(
  'KUNGalgameTopicReply',
  () => {
    const mode = ref<ReplyStorePersist['mode']>('preview')
    const replyDraft = reactive<ReplyStorePersist['replyDraft']>({
      mainContent: ''
    })
    // Floors this draft is replying to (rendered as removable chips above the
    // editor). Their tokens are folded into the body only at publish time.
    const replyReferences = ref<ReplyStorePersist['replyReferences']>([])

    // 「引用」 a floor: add it as a reference chip (deduped by reply id). Kept out
    // of the editor content entirely — no inline header, no caret juggling.
    const addReference = (reference: ReplyReference) => {
      if (replyReferences.value.some((r) => r.replyId === reference.replyId)) {
        return
      }
      replyReferences.value.push(reference)
    }

    const removeReference = (replyId: number) => {
      replyReferences.value = replyReferences.value.filter(
        (r) => r.replyId !== replyId
      )
    }

    // The markdown header prepended to the body on publish: one
    // `[@name](kungal-user:id) [#floor](kungal-reply:id)` per reference. Empty
    // when there are no references.
    const buildReferenceHeader = () =>
      replyReferences.value
        .map(
          (r) =>
            `[@${r.userName}](kungal-user:${r.userId}) [#${r.floor}](kungal-reply:${r.replyId})`
        )
        .join(' ')

    const resetReplyDraft = () => {
      replyDraft.mainContent = ''
      replyReferences.value = []
    }

    return {
      mode,
      replyDraft,
      replyReferences,
      addReference,
      removeReference,
      buildReferenceHeader,
      resetReplyDraft
    }
  },
  {
    persist: true
  }
)
