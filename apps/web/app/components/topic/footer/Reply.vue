<script setup lang="ts">
const { id } = usePersistUserStore()
const { isEdit } = storeToRefs(useTempReplyStore())
const { appendReply } = usePersistKUNGalgameReplyStore()

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

  // Replying to a specific floor: drop a "> 回复 @author #floor" blockquote
  // header into the draft (same shape as migrated replies). The @mention
  // notifies the author, the #quote links the floor, and because the header is a
  // blockquote the editor's trailing empty paragraph lands the caret on the line
  // below — where the reply body goes. Replying to the topic (floor 0) just opens.
  if (props.targetFloor !== 0 && props.targetReplyId) {
    appendReply(
      `> 回复 [@${props.targetUserName}](kungal-user:${props.targetUserId}) [#${props.targetFloor}](kungal-reply:${props.targetReplyId})`
    )
  }

  isEdit.value = true
}
</script>

<template>
  <KunButton variant="flat" @click="handleClickReply">回复</KunButton>
</template>
