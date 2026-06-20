<script setup lang="ts">
const props = defineProps<{
  topicId?: number
  replyId?: number
  targetUserId: number
  dislikeCount: number
  isDisliked: boolean
  // Render as a left-justified labeled row for the ⋯ overflow menu instead of
  // the compact tooltip icon button used inline.
  menu?: boolean
}>()

const { id } = usePersistUserStore()
const isDisliked = ref(props.isDisliked)
const dislikeCount = ref(props.dislikeCount)

const toggleDislike = async () => {
  const result = props.topicId
    ? await kunFetch<string>(`/topic/${props.topicId}/dislike`, {
        method: 'PUT'
      })
    : // Reply branch: same URL-shape fix as Like.vue — avoid
      // `/topic/undefined/reply/dislike` from an undefined topicId.
      await kunFetch<string>(`/topic/0/reply/dislike`, {
        method: 'PUT',
        body: { replyId: props.replyId }
      })

  if (!result) return

  dislikeCount.value += isDisliked.value ? -1 : 1
  useMessage(isDisliked.value ? 10226 : 10225, 'success')
  isDisliked.value = !isDisliked.value
}

const handleClickDislikeThrottled = throttle(toggleDislike, 1007, () =>
  useMessage(10227, 'warn')
)

const handleClickDislike = () => {
  if (!id) {
    useAuthModal().open()
    return
  }
  if (id === props.targetUserId) {
    useMessage(10229, 'warn')
    return
  }
  handleClickDislikeThrottled()
}
</script>

<template>
  <KunButton
    v-if="menu"
    :variant="isDisliked ? 'flat' : 'light'"
    :color="isDisliked ? 'secondary' : 'default'"
    size="sm"
    class-name="w-full justify-start gap-2 whitespace-nowrap"
    @click="handleClickDislike"
  >
    <KunIcon class-name="text-lg" name="lucide:thumbs-down" />
    点踩
    <span v-if="dislikeCount" class="text-default-500 ml-auto">
      {{ dislikeCount }}
    </span>
  </KunButton>

  <KunTooltip v-else text="点踩">
    <KunButton
      :variant="isDisliked ? 'flat' : 'light'"
      :color="isDisliked ? 'secondary' : 'default'"
      :size="dislikeCount ? 'md' : 'lg'"
      class-name="gap-1"
      @click="handleClickDislike"
    >
      <KunIcon class="icon" name="lucide:thumbs-down" />
      <span v-if="dislikeCount">{{ dislikeCount }}</span>
    </KunButton>
  </KunTooltip>
</template>
