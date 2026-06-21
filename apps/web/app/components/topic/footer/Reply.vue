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

  // Replying to a specific floor: drop a plain "@author #floor" header into the
  // draft (no blockquote / 「回复」 wording). The @mention notifies the author,
  // the #quote links the floor, and the editor lands the caret on the line below
  // so the reply body starts there. Replying to the topic (floor 0) just opens.
  if (props.targetFloor !== 0 && props.targetReplyId) {
    appendReply(
      `[@${props.targetUserName}](kungal-user:${props.targetUserId}) [#${props.targetFloor}](kungal-reply:${props.targetReplyId})`
    )
  }

  isEdit.value = true
}
</script>

<template>
  <KunButton variant="flat" @click="handleClickReply">回复</KunButton>
</template>
