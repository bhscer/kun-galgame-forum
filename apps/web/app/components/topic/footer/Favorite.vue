<script setup lang="ts">
const props = defineProps<{
  topicId: number
  favoriteCount: number
  isFavorite: boolean
}>()

const { id } = usePersistUserStore()
const isFavorite = ref(props.isFavorite)
const favoriteCount = ref(props.favoriteCount)
const pending = ref(false)

// The feed card hydrates is-favorite asynchronously (useMyTopicInteractions);
// reflect a late-arriving initial state. Harmless where the prop is settled.
watch(
  () => props.isFavorite,
  (value) => (isFavorite.value = value)
)

// KunReaction flips isFavorite + favoriteCount optimistically before @change;
// we fire the API and undo on failure / when not signed in. `pending` disables
// the control during the request (replaces the old click throttle).
const revert = (next: boolean) => {
  isFavorite.value = !next
  favoriteCount.value += next ? -1 : 1
}

const onChange = async (next: boolean) => {
  if (!id) {
    useAuthModal().open()
    revert(next)
    return
  }
  pending.value = true
  const result = await kunFetch<string>(`/topic/${props.topicId}/favorite`, {
    method: 'PUT'
  })
  pending.value = false
  if (!result) {
    revert(next)
    return
  }
  useMessage(next ? 10230 : 10231, 'success')
}
</script>

<template>
  <KunTooltip text="收藏">
    <KunReaction
      v-model="isFavorite"
      v-model:count="favoriteCount"
      :disabled="pending"
      icon="lucide:heart"
      color="danger"
      label="收藏"
      @change="onChange"
    />
  </KunTooltip>
</template>
