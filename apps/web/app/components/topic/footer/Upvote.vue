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

const upvoteTopic = async () => {
  const res = await useComponentMessageStore().alert(
    '您确定推这个话题吗?',
    '推话题将会消耗您 20 萌萌点, 并给被推者增加 3 萌萌点。'
  )
  if (!res) {
    return
  }

  const result = await kunFetch<string>(
    `/topic/${props.topicId}/upvote`,
    { method: 'PUT' }
  )

  if (result) {
    upvoteCount.value++
    isUpvoted.value = true
    useMessage(10238, 'success')
  }
}

const handleClickUpvote = async () => {
  if (!id) {
    useAuthModal().open()
    return
  }

  if (id === props.targetUserId) {
    useMessage(10241, 'warn')
    return
  }

  if (moemoepoint < 20) {
    useMessage(10242, 'warn')
    return
  }

  await upvoteTopic()
}
</script>

<template>
  <KunButton
    v-if="menu"
    :variant="isUpvoted ? 'flat' : 'light'"
    :color="isUpvoted ? 'secondary' : 'default'"
    size="sm"
    class-name="w-full justify-start gap-2 whitespace-nowrap"
    @click="handleClickUpvote"
  >
    <KunIcon class-name="text-lg" name="lucide:sparkles" />
    推话题
    <span v-if="upvoteCount" class="text-default-500 ml-auto">
      {{ upvoteCount }}
    </span>
  </KunButton>

  <KunTooltip v-else text="推！">
    <KunButton
      :variant="isUpvoted ? 'flat' : 'light'"
      :color="isUpvoted ? 'secondary' : 'default'"
      :size="upvoteCount ? 'md' : 'lg'"
      class-name="gap-1"
      @click="handleClickUpvote"
    >
      <KunIcon class="icon" name="lucide:sparkles" />
      <span v-if="upvoteCount">{{ upvoteCount }}</span>
    </KunButton>
  </KunTooltip>
</template>
