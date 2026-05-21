import { defineStore } from 'pinia'
import { ref } from 'vue'

// Tiny temp store for the galgame comment panel.
//
// Held the target user id across a 3-way pas-de-deux between
// galgame/comment/Container.vue (sets it), Comment.vue (reads + writes
// on reply), and Panel.vue (consumes when sending). Used to live
// alongside resource-publish state in `tempGalgameResource`, but that
// store's resource / rewriteResourceId / isShowPublish fields are all
// gone now (the new LinkEditModal owns its own state), so the comment
// state moved here under its own name to keep the store's purpose
// honest.
export const useTempGalgameCommentStore = defineStore(
  'tempGalgameComment',
  () => {
    const commentToUserId = ref(0)

    const resetGalgameComment = () => {
      commentToUserId.value = 0
    }

    return { commentToUserId, resetGalgameComment }
  },
  { persist: false }
)
