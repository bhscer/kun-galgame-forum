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

  // Replying to a specific floor: drop an @mention of its author (notifies them)
  // + a #quote of the floor into the draft. The mention/quote tokens render as
  // chips in the composer. Replying to the topic itself (floor 0) just opens it.
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
