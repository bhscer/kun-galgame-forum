<script setup lang="ts">
const { id } = usePersistUserStore()
const { isEdit } = storeToRefs(useTempReplyStore())
const { addTarget } = usePersistKUNGalgameReplyStore()

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

  if (props.targetFloor !== 0 && props.targetReplyId) {
    addTarget({
      targetReplyId: props.targetReplyId,
      targetFloor: props.targetFloor,
      targetUserName: props.targetUserName
    })
  }

  isEdit.value = true
}
</script>

<template>
  <KunButton variant="flat" @click="handleClickReply">回复</KunButton>
</template>
