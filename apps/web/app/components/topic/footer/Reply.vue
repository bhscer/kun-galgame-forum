<script setup lang="ts">
const { id } = usePersistUserStore()
const { isEdit } = storeToRefs(useTempReplyStore())
const { addReference } = usePersistKUNGalgameReplyStore()

const props = defineProps<{
  targetUserName: string
  targetUserId: number
  targetFloor: number
  targetReplyId?: number
}>()

const handleClickReply = () => {
  if (!id) {
    useAuthModal().open()
    return
  }

  // Replying to a specific floor: record it as a reference chip (shown above the
  // editor, removable). Its @mention (notifies the author) + #quote (links the
  // floor) tokens are prepended to the body on publish — the editor itself stays
  // an empty composer. Replying to the topic (floor 0) just opens the panel.
  if (props.targetFloor !== 0 && props.targetReplyId) {
    addReference({
      userId: props.targetUserId,
      userName: props.targetUserName,
      replyId: props.targetReplyId,
      floor: props.targetFloor
    })
  }

  isEdit.value = true
}
</script>

<template>
  <KunButton variant="flat" @click="handleClickReply">回复</KunButton>
</template>
