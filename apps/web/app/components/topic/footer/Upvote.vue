<script setup lang="ts">
const props = defineProps<{
  topicId?: number
  targetUserId: number
  upvoteCount: number
  isUpvoted: boolean
  // Render as a left-justified labeled row for the ⋯ overflow menu.
  menu?: boolean
}>()

const { id, moemoepoint } = usePersistUserStore()
const isUpvoted = ref(props.isUpvoted)
const upvoteCount = ref(props.upvoteCount)
const pending = ref(false)

// Shared: self / moemoepoint checks → confirm dialog → API. Returns true once
// the upvote actually lands. Upvote is one-way (no undo) and costs 20 萌萌点.
const confirmAndUpvote = async (): Promise<boolean> => {
  if (id === props.targetUserId) {
    useMessage(10241, 'warn')
    return false
  }
  if (moemoepoint < 20) {
    useMessage(10242, 'warn')
    return false
  }
  const ok = await useComponentMessageStore().alert(
    '您确定推这个话题吗?',
    '推话题将会消耗您 20 萌萌点, 并给被推者增加 3 萌萌点。'
  )
  if (!ok) return false

  pending.value = true
  const result = await kunFetch<string>(`/topic/${props.topicId}/upvote`, {
    method: 'PUT'
  })
  pending.value = false
  if (!result) return false

  useMessage(10238, 'success')
  return true
}

// Menu (KunButton) path: imperative — set state only after a confirmed upvote.
const handleClickUpvote = async () => {
  if (!id) {
    useAuthModal().open()
    return
  }
  if (isUpvoted.value || pending.value) return
  if (await confirmAndUpvote()) {
    upvoteCount.value++
    isUpvoted.value = true
  }
}

// KunReaction path: it already flipped to upvoted; keep it on success, undo otherwise.
const revert = () => {
  isUpvoted.value = false
  upvoteCount.value--
}
const onChange = async (next: boolean) => {
  if (!next) return
  if (!id) {
    useAuthModal().open()
    revert()
    return
  }
  if (!(await confirmAndUpvote())) revert()
}
</script>

<template>
  <KunButton
    v-if="menu"
    :variant="isUpvoted ? 'flat' : 'light'"
    :color="isUpvoted ? 'secondary' : 'default'"
    size="sm"
    class-name="w-full justify-start gap-2 whitespace-nowrap"
    :disabled="pending"
    @click="handleClickUpvote"
  >
    <KunIcon class-name="text-lg" name="lucide:sparkles" />
    推话题
    <span v-if="upvoteCount" class="text-default-500 ml-auto">
      {{ upvoteCount }}
    </span>
  </KunButton>

  <KunTooltip v-else text="推话题">
    <KunReaction
      v-model="isUpvoted"
      v-model:count="upvoteCount"
      :disabled="isUpvoted || pending"
      icon="lucide:sparkles"
      color="warning"
      label="推话题"
      @change="onChange"
    />
  </KunTooltip>
</template>
